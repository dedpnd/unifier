package worker

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dedpnd/unifier/internal/models"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func Start(ctx context.Context, kafkaURL string, wrkConfig workerEntity, lg *zap.Logger) error {
	var r *kafka.Reader
	var p *kafka.Conn

	lg.Info("Worker start", zap.String("ID", wrkConfig.ID))

	// Создаем kafka consumer
	r = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		GroupID: wrkConfig.ID,
		Topic:   wrkConfig.Config.TopicFrom,
	})

	// Создаем kafka producer
	var err error
	p, err = kafka.DialLeader(context.Background(), "tcp", kafkaURL, wrkConfig.Config.TopicTo, 0)
	if err != nil {
		return fmt.Errorf("worker:%v - failed create producer: %w", wrkConfig.ID, err)
	}
	// Вычитываем сообщения
	for {
		select {
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				return fmt.Errorf("worker:%v - failed read message: %w", wrkConfig.ID, err)
			}

			// Фильтруем событие по регулярному выражению
			strMsg := string(msg.Value)
			matched, err := regexp.Match(wrkConfig.Config.Filter.Regexp, []byte(strMsg))
			if err != nil {
				return fmt.Errorf("worker:%v - failed filter message: %w", wrkConfig.ID, err)
			}

			// Преобразум сообщения для удобства разбора
			if matched {
				var uniferEvents = make(map[string]interface{})
				var pEvent map[string]interface{}

				if err := json.Unmarshal([]byte(strMsg), &pEvent); err != nil {
					return fmt.Errorf("worker:%v - invalid JSON parse: %w", wrkConfig.ID, err)
				}

				// Вычисляем уникальных идентификатор для записи
				hex := calculateHash(pEvent, wrkConfig.Config.EntityHash)
				uniferEvents["entity"] = hex

				// Унификация полей
				err = unificationFields(pEvent, wrkConfig.Config.Unifier, &uniferEvents)
				if err != nil {
					lg.Error(err.Error())
				}

				// Допольнительная обработка
				err = extraProcess(wrkConfig.Config.ExtraProcess, &uniferEvents)
				if err != nil {
					lg.Error(err.Error())
				}

				buf, err := json.Marshal(uniferEvents)
				if err != nil {
					return fmt.Errorf("worker:%v - failed stringify message: %w", wrkConfig.ID, err)
				}

				_, err = p.WriteMessages(kafka.Message{Value: buf})
				if err != nil {
					return fmt.Errorf("worker:%v - failed to write messages: %w", wrkConfig.ID, err)
				}
			}
		case <-wrkConfig.Stop:
			lg.Info("Worker stop", zap.String("ID", wrkConfig.ID))

			err := r.Close()
			if err != nil {
				return fmt.Errorf("worker:%v - failed close consumer: %w", wrkConfig.ID, err)
			}

			err = p.Close()
			if err != nil {
				return fmt.Errorf("worker:%v - failed close producer: %w", wrkConfig.ID, err)
			}

			return nil
		}
	}
}

func extraProcess(cfgExtraProcess []models.ExtraProcess, uEvent *map[string]interface{}) error {
	for _, ep := range cfgExtraProcess {
		switch ep.Func {
		case "__if":
			r := __if(*uEvent, ep.Args)
			(*uEvent)[ep.To] = r
		case "__stringConstant":
			r := __stringConstant(ep.Args)
			(*uEvent)[ep.To] = r
		default:
			return fmt.Errorf("unknown func: %v", ep.Func)
		}
	}

	return nil
}

func unificationFields(event map[string]interface{}, cfgUnifier []models.Unifier, uEvent *map[string]interface{}) error {
	for _, u := range cfgUnifier {
		v, found := event[u.Expression]
		if found {
			switch u.Type {
			// TODO: Логировать когда преобразование не получилось
			case "string":
				v, ok := v.(string)
				if ok {
					(*uEvent)[u.Name] = v
				}
			case "int":
				vv, ok := v.(int)
				if ok {
					(*uEvent)[u.Name] = vv
					return nil
				}

				v, ok := v.(string)
				if !ok {
					return fmt.Errorf("failed int parse to string: %v", v)
				}

				i, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("failed int parse: %w", err)
				}

				(*uEvent)[u.Name] = i
			case "timestamp":
				v, ok := v.(string)
				if ok {
					v, err := time.Parse(time.RFC3339, v)

					if err != nil {
						return fmt.Errorf("failed date parse: %w", err)
					}

					(*uEvent)[u.Name] = v.Format(time.RFC3339)
				}
			default:
				return fmt.Errorf("unknown type: %v", u.Type)
			}
		}
	}

	return nil
}

func calculateHash(event map[string]interface{}, cfgEntHash []string) string {
	strHash := ""
	for _, eh := range cfgEntHash {
		v, found := event[eh]
		if found {
			s, ok := v.(string)
			if ok {
				strHash += strings.TrimSpace(s)
			}
		}
	}

	// Вычисляем хэш
	hash := md5.Sum([]byte(strHash))
	hexStr := hex.EncodeToString(hash[:])

	return hexStr
}

// Function fot extra process!
//
//nolint:stylecheck // This legal name
func __if(uniEvent map[string]interface{}, args string) string {
	aSlice := strings.Split(args, ",")
	for i, v := range aSlice {
		aSlice[i] = strings.TrimSpace(v)
	}

	var field, stmn, result string = aSlice[0], aSlice[1], aSlice[2]

	v, ok := uniEvent[field]
	if ok {
		if v == stmn {
			return result
		}
	}

	return ""
}

//nolint:stylecheck // This legal name
func __stringConstant(args string) string {
	aSlice := strings.Split(args, ",")
	for i, v := range aSlice {
		aSlice[i] = strings.TrimSpace(v)
	}

	var c string = aSlice[0]
	return c
}

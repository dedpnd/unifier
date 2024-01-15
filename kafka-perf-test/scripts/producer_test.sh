#!/usr/bin/env bash

./generate_endpoints.sh

while [[ true ]]
do
    bin/kafka-producer-perf-test.sh --producer.config ./config/producer.properties --print-metrics --throughput ${PRODUCER_THROUGHPUT} --num-records ${NUM_RECORDS} --topic ${KAFKA_TOPIC} --payload-file example.json
    sleep ${TEST_INTERVAL_SECONDS}    
done

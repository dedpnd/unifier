BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS users(	
	ID    SERIAL UNIQUE PRIMARY KEY NOT NULL,
	Login VARCHAR(255) UNIQUE NOT NULL,
	Hash  VARCHAR(1000) NOT NULL
);

CREATE TABLE IF NOT EXISTS rules(	
	ID SERIAL PRIMARY KEY NOT NULL,
  Rule JSON NOT NULL,
	Owner INT NULL,
	CONSTRAINT fk_users
      FOREIGN KEY(Owner) 
				REFERENCES users(ID)
				ON DELETE SET NULL
);

INSERT INTO rules
("rule", "owner")
VALUES('{"topicFrom":"events","filter":{"regexp":"\"dstHost.ip\": \"10.10.10.10\""},"entityHash":["srcHost.ip","dstHost.port"],"unifier":[{"name":"id","type":"string","expression":"auditEventLog"},{"name":"date","type":"timestamp","expression":"datetime"},{"name":"ipaddr","type":"string","expression":"srcHost.ip"},{"name":"category","type":"string","expression":"cat"}],"extraProcess":[{"func":"__if","args":"category, /Host/Connect/Host/Accept, high","to":"category"},{"func":"__stringConstant","args":"test","to":"customString1"}],"topicTo":"test"}'::json, NULL);

COMMIT;
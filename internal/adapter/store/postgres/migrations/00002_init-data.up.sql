-- password is password
INSERT INTO users
(login, hash)
VALUES('admin', '$2a$10$fvuwEbdImMWCPGzjuJ7pbOAQnZ/e9VyrVK60ComfJiEJvgkORJTci'); 

INSERT INTO rules
("rule", "owner")
VALUES('{"topicFrom":"events","filter":{"regexp":"\"dstHost.ip\": \"10.10.10.10\""},"entityHash":["srcHost.ip","dstHost.port"],"unifier":[{"name":"id","type":"string","expression":"auditEventLog"},{"name":"date","type":"timestamp","expression":"datetime"},{"name":"ipaddr","type":"string","expression":"srcHost.ip"},{"name":"category","type":"string","expression":"cat"}],"extraProcess":[{"func":"__if","args":"category, /Host/Connect/Host/Accept, high","to":"category"},{"func":"__stringConstant","args":"test","to":"customString1"}],"topicTo":"test"}'::json, 1);

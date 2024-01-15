CREATE TABLE IF NOT EXISTS users(	
	ID    SERIAL UNIQUE PRIMARY KEY NOT NULL,
	Login  VARCHAR(255) UNIQUE NOT NULL,
	Hash  VARCHAR(1000) NOT NULL
);

CREATE TABLE IF NOT EXISTS rules(	
	ID SERIAL PRIMARY KEY,
  Rule JSON,
	Owner INT,
	CONSTRAINT fk_users
      FOREIGN KEY(Owner) 
				REFERENCES users(ID)
				ON DELETE SET NULL
);
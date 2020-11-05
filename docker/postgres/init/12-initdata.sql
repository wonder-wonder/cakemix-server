\c cakemix;

-- System Admin (default user)
-- Email:root@localhost Pass:cakemix
INSERT INTO username VALUES('ujafzavrqkqthqe54', 'root');
INSERT INTO auth VALUES('ujafzavrqkqthqe54',	'root@localhost',	'DBerQ+J0ywuKJ+sSXHx9Y/5L4qxhL/275f3d70YjINmgj5ftoNL9yu42aujAEpUUTYiZZUpdqojuhRj7ry3ISQ==',	'nEmGLz2FIqoOJAsN');
INSERT INTO profile VALUES('ujafzavrqkqthqe54','','',1,'','ja');

-- System Admin Team
INSERT INTO username VALUES('tqssoagvfvlg3mky2', 'systemadmin');
INSERT INTO profile VALUES('tqssoagvfvlg3mky2','','',1,'','ja');
INSERT INTO teammember VALUES('tqssoagvfvlg3mky2', 'ujafzavrqkqthqe54', 0, 1);

-- Default Environments
-- Default tag
INSERT INTO tag VALUES (0,'notag');
-- Root folder
INSERT INTO folder VALUES('fwk6al7nyj4qdufaz','tqssoagvfvlg3mky2','','',1,1,1,'ujafzavrqkqthqe54');
-- User folder
INSERT INTO folder VALUES('fdahpbkboamdbgnua','tqssoagvfvlg3mky2','fwk6al7nyj4qdufaz','User',1,1,1,'ujafzavrqkqthqe54');

-- Admin folder
INSERT INTO folder VALUES('fhfprvdljyczssis7','ujafzavrqkqthqe54','fdahpbkboamdbgnua','root',0,1,1,'ujafzavrqkqthqe54');

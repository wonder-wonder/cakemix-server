\c cakemix;

-- Test user
INSERT INTO username VALUES('ujuxj7nrznlg655jt', 'test');
INSERT INTO auth VALUES('ujuxj7nrznlg655jt',	'test@localhost',	'CgL73Jxt1TcOIemWTZye7FkVNIz2kF+5cLG4RzAEsxhpow98YJGQr4yaK2IpbT5gwk097EGNcu/hi90nCZrN8w==',	'GlORyqKbYphmrv9I');
INSERT INTO profile VALUES('ujuxj7nrznlg655jt','test','','',1595073002,'','ja');
INSERT INTO folder VALUES('fde2uhiehs25fahxk','ujuxj7nrznlg655jt','fdahpbkboamdbgnua','test',0,1595073002,1595073002,'ujuxj7nrznlg655jt');

INSERT INTO teammember VALUES('tqssoagvfvlg3mky2', 'ujuxj7nrznlg655jt', 0, 1595073002);

insert into tag values (1,'tag1');
insert into document values ('dotyrr3hpdeyvpwhy','ujuxj7nrznlg655jt','fde2uhiehs25fahxk','title',2,1595073002,1595073002,'ujuxj7nrznlg655jt',0);
insert into documentrevision values ('dotyrr3hpdeyvpwhy','',1);
insert into documentrevision values ('dotyrr3hpdeyvpwhy','This is a test.',1595073002);
-- INSERT INTO banks(bank_name,bank_abbreviation,is_active) VALUES('HDFC','HDFC',1);
-- INSERT INTO bank_users (bank_id,user_id) VALUES(1,7)
-- INSERT INTO clients (client_name,client_email,balance,is_active,bank_id,verification_status) VALUES('TCS','TCS@gmail.com',1200,1,1,"Approved")
-- INSERT INTO client_users(client_id,user_id) VALUES(2,10)
-- UPDATE clients SET balance = 100000 WHERE id =1;
-- DELETE From roles WHERE roles.id >3

DROP TABLE payments;
DROP TABLE payment_requests;
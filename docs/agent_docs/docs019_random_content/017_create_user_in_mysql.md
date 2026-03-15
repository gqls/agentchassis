kubectl -n ai-persona-system exec -it auth-service-656db99bb6-lcsvm -- wget -qO- \
http://localhost:8081/api/v1/auth/register \
--post-data='{"email":"uk@websy.uk","password":"AdminUser$%^PW123!","client_id":"demo_client","role":"admin"}' \
--header='Content-Type: application/json' 2>&1
{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOWM3YTFjOWUtNjU4Mi00NWM1LWFlYmQtNDE4NGI1YzkyZDM4IiwiZW1haWwiOiJ1a0B3ZWJzeS51ayIsImNsaWVudF9pZCI6ImRlbW9fY2xpZW50Iiwicm9sZSI6InVzZXIiLCJ0aWVyIjoiZnJlZSIsImlzcyI6ImFpLXBlcnNvbmEtc3lzdGVtIiwic3ViIjoiOWM3YTFjOWUtNjU4Mi00NWM1LWFlYmQtNDE4NGI1YzkyZDM4IiwiZXhwIjoxNzczNTgwNDYyLCJuYmYiOjE3NzM1NzY4NjIsImlhdCI6MTc3MzU3Njg2MiwianRpIjoiMTc3MzU3Njg2MiJ9.sUNnTZISCvREP5ieS0Eob0nC33Zq8Ouzf6nlCP2yK-4","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOWM3YTFjOWUtNjU4Mi00NWM1LWFlYmQtNDE4NGI1YzkyZDM4Iiwic3ViIjoiOWM3YTFjOWUtNjU4Mi00NWM1LWFlYmQtNDE4NGI1YzkyZDM4IiwiZXhwIjoxNzc0MTgxNjYyLCJpYXQiOjE3NzM1NzY4NjIsImp0aSI6InJlZnJlc2hfMTc3MzU3Njg2MiJ9.SWmNdUiUSlKpvrAcUQEK2j1ZZhEeTPmZZlqy5UEKiZY","token_type":"Bearer","expires_in":3600,"user":{"id":"9c7a1c9e-6582-45c5-aebd-4184b5c92d38","email":"uk@websy.uk","client_id":"demo_client","role":"user","tier":"free","email_verified":false,"permissions":null}}

Registered, but the token came back with role: "user" and tier: "free" — the auth service ignored the role field. Promote it in MySQL:
mysql -h rs17.uk-noc.com -u catalogu_personae -p"PpC47410423123!" --skip-ssl catalogu_vectordb_chassis \
-e "UPDATE users SET role = 'admin', subscription_tier = 'enterprise' WHERE email = 'uk@websy.uk';"

to get into mysql start up a pod
ant@ant-XPS-15-9500:~/projects/agentchassis$ kubectl -n ai-persona-system run mysql-check --rm -it --image=postgres:16-alpine -- /bin/sh
install mysql client with apk
apk add --no-cache mysql-client

mysql -h rs17.uk-noc.com -u catalogu_personae -p"PpC47410423123!" --skip-ssl catalogu_vectordb_chassis \
> -e "UPDATE users SET role = 'admin', subscription_tier = 'enterprise' WHERE email = 'uk@websy.uk';"

+------------------+-------------+--------------------------------------------------------------+-------+-------------+-------------------+-----------+----------------+---------------------+---------------------+
| id               | email       | password_hash                                                | role  | client_id   | subscription_tier | is_active | email_verified | created_at          | updated_at          |
+------------------+-------------+--------------------------------------------------------------+-------+-------------+-------------------+-----------+----------------+---------------------+---------------------+
| 9c7a1c9e-6582-45 | uk@websy.uk | $2a$10$DJWgeCrPk3RAn196HPwoz.p8a7RgtrZ7n71sX2kel3ChU9Lq57E4S | admin | demo_client | enterprise        |         1 |              0 | 2026-03-15 12:14:23 | 2026-03-15 12:18:26 |
+------------------+-------------+--------------------------------------------------------------+-------+-------------+-------------------+-----------+----------------+---------------------+---------------------+

Then get a fresh token with the updated role:
kubectl -n ai-persona-system exec -it auth-service-656db99bb6-lcsvm -- wget -qO- \
http://localhost:8081/api/v1/auth/login \
--post-data='{"email":"uk@websy.uk","password":"AdminUser$%^PW123!"}' \
--header='Content-Type: application/json' 2>&1

The new JWT should have role: "admin" in the claims, which will pass the AdminOnly() middleware. Once you have that token, test the admin API:
TOKEN="the_new_access_token"
curl -s http://core-manager.ai-persona-system.svc.cluster.local:8088/api/v1/admin/sites \
-H "Authorization: Bearer $TOKEN" | python3 -m json.tool | head -30

Delete and re-register with a simpler password:
mysql -h rs17.uk-noc.com -u catalogu_personae -p"PpC47410423123!" --skip-ssl catalogu_vectordb_chassis \
-e "DELETE FROM user_profiles WHERE user_id = (SELECT id FROM users WHERE email = 'uk@websy.uk'); DELETE FROM auth_tokens WHERE user_id = (SELECT id FROM users WHERE email = 'uk@websy.uk'); DELETE FROM users WHERE email = 'uk@websy.uk';"

Then register again with something shell-safe:
kubectl -n ai-persona-system exec -it auth-service-656db99bb6-lcsvm -- wget -qO- \
http://localhost:8081/api/v1/auth/register \
--post-data='{"email":"uk@websy.uk","password":"AdminPass2026xyz","client_id":"demo_client"}' \
--header='Content-Type: application/json' 2>&1
{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTkzZjZiMjktMzE2MS00ZTc4LTllYWYtNTc2NGI4Yjc2NjhlIiwiZW1haWwiOiJ1a0B3ZWJzeS51ayIsImNsaWVudF9pZCI6ImRlbW9fY2xpZW50Iiwicm9sZSI6InVzZXIiLCJ0aWVyIjoiZnJlZSIsImlzcyI6ImFpLXBlcnNvbmEtc3lzdGVtIiwic3ViIjoiMTkzZjZiMjktMzE2MS00ZTc4LTllYWYtNTc2NGI4Yjc2NjhlIiwiZXhwIjoxNzczNTgxMzAxLCJuYmYiOjE3NzM1Nzc3MDEsImlhdCI6MTc3MzU3NzcwMSwianRpIjoiMTc3MzU3NzcwMSJ9.2RR2fsQ_CBBFoqpCVsm-vwYK571MGSpozhhs0n1CbrY","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTkzZjZiMjktMzE2MS00ZTc4LTllYWYtNTc2NGI4Yjc2NjhlIiwic3ViIjoiMTkzZjZiMjktMzE2MS00ZTc4LTllYWYtNTc2NGI4Yjc2NjhlIiwiZXhwIjoxNzc0MTgyNTAxLCJpYXQiOjE3NzM1Nzc3MDEsImp0aSI6InJlZnJlc2hfMTc3MzU3NzcwMSJ9.pIHMGlFs9oN7jhltT6rHugtCruo5Z3gfTu7is-NQ2Ok","token_type":"Bearer","expires_in":3600,"user":{"id":"193f6b29-3161-4e78-9eaf-5764b8b7668e","email":"uk@websy.uk","client_id":"demo_client","role":"user","tier":"free","email_verified":false,"permissions":null}}

Then promote and login:
mysql -h rs17.uk-noc.com -u catalogu_personae -p"PpC47410423123!" --skip-ssl catalogu_vectordb_chassis \
-e "UPDATE users SET role = 'admin', subscription_tier = 'enterprise' WHERE email = 'uk@websy.uk';"

kubectl -n ai-persona-system exec -it auth-service-656db99bb6-hmcm8 -- wget -qO- \
http://localhost:8081/api/v1/auth/login \
--post-data='{"email":"uk@websy.uk","password":"AdminPass2026xyz","client_id":"demo_client"}' \
--header='Content-Type: application/json' 2>&1


-- fix the UUID issue

mysql -h rs17.uk-noc.com -u catalogu_personae -p"PpC47410423123!" --skip-ssl catalogu_vectordb_chassis -e "
ALTER TABLE auth_tokens DROP FOREIGN KEY auth_tokens_ibfk_1;
ALTER TABLE projects DROP FOREIGN KEY projects_ibfk_1;
ALTER TABLE subscriptions DROP FOREIGN KEY subscriptions_ibfk_1;
ALTER TABLE user_permissions DROP FOREIGN KEY user_permissions_ibfk_1;
ALTER TABLE user_profiles DROP FOREIGN KEY user_profiles_ibfk_1;
DELETE FROM user_profiles;
DELETE FROM auth_tokens;
DELETE FROM user_permissions;
DELETE FROM subscriptions;
DELETE FROM projects;
DELETE FROM users;
ALTER TABLE users MODIFY COLUMN id CHAR(36) NOT NULL;
ALTER TABLE auth_tokens MODIFY COLUMN user_id CHAR(36) NOT NULL;
ALTER TABLE projects MODIFY COLUMN user_id CHAR(36) NOT NULL;
ALTER TABLE subscriptions MODIFY COLUMN user_id CHAR(36) NOT NULL;
ALTER TABLE user_permissions MODIFY COLUMN user_id CHAR(36) NOT NULL;
ALTER TABLE user_profiles MODIFY COLUMN user_id CHAR(36) NOT NULL;
ALTER TABLE auth_tokens ADD CONSTRAINT auth_tokens_ibfk_1 FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE projects ADD CONSTRAINT projects_ibfk_1 FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_ibfk_1 FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE user_permissions ADD CONSTRAINT user_permissions_ibfk_1 FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE user_profiles ADD CONSTRAINT user_profiles_ibfk_1 FOREIGN KEY (user_id) REFERENCES users(id);
"


get fresh token
kubectl -n ai-persona-system exec -it $(kubectl -n ai-persona-system get pod -l app=auth-service -o jsonpath='{.items[0].metadata.name}') -- wget -qO- \
http://localhost:8081/api/v1/auth/login \
--post-data='{"email":"uk@websy.uk","password":"AdminPass2026xyz"}' \
--header='Content-Type: application/json' 2>&1


TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOTU1ZDQ5MTMtYjI2OS00ODg5LWI5MWEtNzIxMzY0Njk0YTBlIiwiZW1haWwiOiJ1a0B3ZWJzeS51ayIsImNsaWVudF9pZCI6ImRlbW9fY2xpZW50Iiwicm9sZSI6ImFkbWluIiwidGllciI6ImVudGVycHJpc2UiLCJpc3MiOiJhaS1wZXJzb25hLXN5c3RlbSIsInN1YiI6Ijk1NWQ0OTEzLWIyNjktNDg4OS1iOTFhLTcyMTM2NDY5NGEwZSIsImV4cCI6MTc3MzU5Nzc4MiwibmJmIjoxNzczNTk0MTgyLCJpYXQiOjE3NzM1OTQxODIsImp0aSI6IjE3NzM1OTQxODIifQ.654VlA2lJ5TgbneDOhqAJIoowXU37SHXWaMe4bS1neE"

curl -s http://localhost:8088/api/v1/admin/sites -H "Authorization: Bearer $TOKEN" | python3 -m json.tool | head -40
Extra data: line 1 column 5 (char 4)

debug
Raw response first:
curl -s http://localhost:8088/api/v1/admin/sites -H "Authorization: Bearer $TOKEN"
# CREATE DOMAIN 
aws es --endpoint-url=http://localhost.localstack.cloud:4566 --region us-east-1 create-elasticsearch-domain --domain-name my-domain
# "Endpoint": "my-domain.us-east-1.es.localhost.localstack.cloud:4566",
# aws es --endpoint-url=http://localhost.localstack.cloud:4566 --region us-east-1  describe-elasticsearch-domain --domain-name my-domain | jq ".DomainStatus.Processing"

# CREATE INDEX INSIDE DOMAIN 
curl -X PUT my-domain.us-east-1.es.localhost.localstack.cloud:4566/my-index
# to sshow sstatus of index 
# http://my-domain.us-east-1.es.localhost.localstack.cloud:4566/my-index

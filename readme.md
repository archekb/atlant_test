Статистика с балансера: http://localhost:8080/stats
gRPC [protobuf]:  localhost:50051 (Reflection not implemented)

В docker-compose сделал два сервиса, чтобы обращаться к ним по имени сервиса из балансировщика. По хорошему нужно было бы использовать scale, но тогда на имя сервиса отдается два ip и нужно делать дополнительную логику для формирования конфигурации балансировщика.

Proto файл собирается через api v1.

Нюансы:
1) Нумерация страниц начинается с 0 (для совместимости с дефолтным значение protobuf)
2) В задании сказано про сортировку, но нет указаний про фильтры, сортировку сделал, а фильтры - нет
3) Цены приводятся к целому числу, чтобы избежать аномалий при работе с float, как и отдаются в результирующем ответе метода List
4) В ответе на метод List в paging добавляется параметр page_cont, который отображает текущее количество страниц в базе
5) Так как в задании нет указаний как необходимо отвечать на Fetch, я решил что при выполнении запроса Fetch вызывающей стороне важно знать, был ли обработан запрос. По этому Fetch реализован прямым вызовом, а не добавлением URL в очередь на обработку.
6) Балансировка происходит на основе source адреса вызывающего, за счет этого, чтобы увидеть балансировку в действии, необходимо делать запросы с разных ip

gRPC cli:
./grpc_cli call localhost:50051 Fetch "url: 'http://localhost/test.csv'" --protofiles=product.proto --proto_path=[proto_path] --noremotedb
./grpc_cli call localhost:50051 List "paging: { result_per_page:25 } sort: [{}]" --protofiles=product.proto --proto_path=[proto_path] --noremotedb
./grpc_cli call localhost:50051 List "paging: { page_number: 1, result_per_page:25 } sort: [{sortby: 1, dsc: true}]" --protofiles=product.proto --proto_path=[proto_path] --noremotedb

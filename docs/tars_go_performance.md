## Bench environment

- Bench Machine type: 4  core /8 thread CPU  3.3Ghz，16G memory

- Bench Logic: The client carries a certain amount of data to the server, and the server returns it to the client as it is.

- Server single process, multiple clients initiate bench.

## Bench result

| framwork       | TPS(10 byte) | TPS(128 byte) | TPS(256 byte) |
| -------------- | ------------ | ------------- | ------------- |
| TARS(C++)      | 617163       | 390686        | 280637        |
| TARS(JAVA)     | 430725       | 384113        | 279531        |
| TARS(NODEJS)   | 158888       | 158139        | 157334        |
| TARS(GO)       | 596795       | 386024        | 276458        |
| TARS(PHP)      | 168745       | 169953        | 168617        |
| Spring   Cloud | 160114       | 157010        | 156830        |
| gRPC(C++）     | 89351        | 86132         | 81630         |
| gRPC(GO)       | 106345       | 100599        | 99684         |
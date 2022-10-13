# fluent-bit-pgsql-go

Fluent Bit Plugin for PostgreSQL

## Usage

pgsql.conf
```
[SERVICE]
  flush 5
  plugins_file ./plugins.conf

[INPUT]
  name winstat
  tag  event-log

[OUTPUT]
  name   pgsql-go
  match  event*
  dsn    <YOUR DATABASE CONNECTION STRING>
  table  events
```

plugins.conf
```
[PLUGINS]
    Path c:/dev/tools/fluent-bit/bin/pgsql-go.dll

```

## Requirements

* fluent-bit

## Installation

```
$ make
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)

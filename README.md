# QBot - A slack bot for unintrusive QnA handling

QBot was written to adress mainly two problems. Intrusive questions and the asking of already answered questions.


![Intrusive Questions](https://img.devrant.com/devrant/rant/r_186393_9yzn5.jpg)


## Getting Started

To deploy QBot, simply clone the repo, populate the fields in the [config.yaml](example_config.yaml) file and run `go run .` in the root directory of the repository.

### Pre-requisites

A Slack bot API token needs to be created. This can be done over at [Slack's official website](https://api.slack.com/).

A Database also needs to be setup. QBot is utilizing the [GORM library](github.com/jinzhu/gorm) for database modeling. Which means that it's nativeley supporting MySQL, Postgres and sqlite3. For MySQL databases, populate the database fields in the [config.yaml](example_config.yaml) as following.

```yaml
database:
  type: "mysql"        # Dialect e.g. sqlite3
  database: "qbot"     # Name of database
  user: "root"         # Username
  password: "qbot"     # Password
```

### Supported commands

![qbot commands](https://imgur.com/17MfKAM)

## Built With

* [GORM library](github.com/jinzhu/gorm) - Object relational mapping library
* [Slack](github.com/nlopes/slack) - Slack API wrapper in Go
* [Zerolog](github.com/rs/zerolog/log) - For STDOUT logging


## Authors

* **Martin Granström** - *Initial work* - [grvlle](https://github.com/grvlle)

See also the list of [contributors](https://github.com/your/project/contributors) who participated in this project.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

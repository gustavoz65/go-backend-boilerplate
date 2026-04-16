// Atlas configuration for the backend boilerplate.

variable "db_user" {
  type    = string
  default = "root"
}

variable "db_pass" {
  type    = string
  default = "password"
}

variable "db_host" {
  type    = string
  default = "localhost"
}

variable "db_port" {
  type    = string
  default = "3306"
}

variable "db_name" {
  type    = string
  default = "boilerplate"
}

env "local" {
  src = "file://internal/schema/schema.hcl"
  url = "mysql://${var.db_user}:${var.db_pass}@${var.db_host}:${var.db_port}/${var.db_name}?parseTime=true"
  dev = "mysql://${var.db_user}:${var.db_pass}@${var.db_host}:${var.db_port}/boilerplate_dev?parseTime=true"

  migration {
    dir = "file://internal/database/migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "prod" {
  src = "file://internal/schema/schema.hcl"
  url = "mysql://${var.db_user}:${var.db_pass}@${var.db_host}:${var.db_port}/${var.db_name}?parseTime=true"
  dev = "mysql://${var.db_user}:${var.db_pass}@${var.db_host}:${var.db_port}/boilerplate_dev?parseTime=true"

  migration {
    dir = "file://internal/database/migrations"
  }
}

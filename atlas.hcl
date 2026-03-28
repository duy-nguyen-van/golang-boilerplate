data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "ariga.io/atlas-provider-gorm@v0.6.0",
    "load",
    "--path", "./internal/models",
    "--dialect", "postgres"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/18/dev"
  migration {
    dir = "file://cmd/migrations/sql"
    exclude = ["*.base_models[type=table]"]
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

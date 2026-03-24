target "auth" {
  secret = [
    {
      id = "github_token"
      type = "env"
      env = "GH_APP_TOKEN"
    }
  ]
}

target "_main" { inherits = [ "auth" ] }
target "default" {
  inherits = [ "_main" ]
  tags = [ "phortizo" ]
}

group "cd" { targets = ["_main"] }

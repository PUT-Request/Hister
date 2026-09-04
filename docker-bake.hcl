group "default" {
  targets = ["release"]
}

group "images" {
  targets = ["release", "root", "debug"]
}

target "_common" {
  context    = "."
  dockerfile = "Dockerfile"
}

target "release" {
  inherits = ["_common"]
  target   = "release"
}

target "root" {
  inherits = ["_common"]
  target   = "root"
}

target "debug" {
  inherits = ["_common"]
  target   = "debug"
}

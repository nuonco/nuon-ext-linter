package nuon

deny contains msg if {
  some resource in input.plan.resource_changes
  resource.mode == "managed"
  msg := "read only"
}

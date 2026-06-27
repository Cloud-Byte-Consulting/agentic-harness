---
id: business-analyst
title: Business Analyst
description: "Describe what this Pod tracks and why it exists."
owners:
  - alias: your-alias
    role: owner
scope:
  systems:
    - github
    - ado
routing:
  defaultRouter: /platform-router
  allowedRouters:
    - /platform-router
    - /team-beta-router
    - /team-alpha-router
writePolicy:
  defaultMode: read-only
  requiresMutationGate: true
---

# Pod Definition

## Intent

Write a short mission statement for this Pod.

## Success Signals

1. Add measurable outcome 1.
2. Add measurable outcome 2.

## Out of Scope

List what this Pod should ignore.

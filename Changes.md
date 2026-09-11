## 0.1.0  2026-09-11

 * First tagged release. Until now consumers have pulled this module at a
   pseudo-version from `master`; from here on releases are tagged `vX.Y.Z` for
   the root module and `gin/vX.Y.Z` for the Gin integration.
 * Struct schemas now carry a `required` array inferred from the Go declaration:
   non-pointer fields without `omitempty`/`omitzero` are required, pointer and
   `omitempty` fields are optional, and `openapi:",required"` /
   `openapi:",optional"` override the inference. Fields promoted from an embedded
   pointer are never required, and an outer field that shadows a promoted field
   decides for itself. (#91)
 * Added `Document.SchemaComponent(fqn, model)` to register any model as a schema
   component after the fact.
 * Gin `Call` gained `WithErrorComponent()`, which registers the default
   `ErrorResponse` and every custom error model under `components/schemas` and
   references them from the `default` response instead of inlining them in every
   operation. `WithComponents()` now enables request, response, and error
   components. (#92)
 * Registering a slice type as a component (a `[]Pet` response with
   `WithResponseComponent()`) now registers the element and renders an array of
   `$ref`, instead of a component named `.`. Pointer types register under the
   element type's name.
 * Field-level Go doc comments now reach the generated schema properties and
   parameters. A trailing line comment on a field is used when there is no doc
   comment above it. (#93)

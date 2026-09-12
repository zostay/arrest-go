## 0.2.0  2026-09-12

This release is about what it costs a consumer to compile against arrest-go
(#99). A project that added arrest-go saw its cold build go from 7s to 24s; the
first cut brings that to about 18s, and the rest is tracked on the epic.

 * **Breaking:** the `libopenapi` root package is no longer imported. It
   brought the Swagger 2, Arazzo, overlay and what-changed packages — ten
   packages this library never used — into every consumer's build, about 12
   CPU-seconds and 4 seconds of a cold build's critical path. `Document.OpenAPI`
   is gone; render with `doc.Render()` instead of `doc.OpenAPI.Render()`.
   `Document.DataModel` is now an `arrest.DocumentModel` with the same `Model`
   and `Index` fields, so `doc.DataModel.Model` is unchanged. `NewDocumentFrom`
   takes a `*v3.Document`. (#100)
 * The order of `components.schemas` is now deterministic: child components
   are registered in name order rather than a Go map's iteration order, and
   `Render()` (and so `Refresh()`) puts the schema and security scheme
   components in name order however they were declared. A document
   regenerated from unchanged handlers renders byte-for-byte the same, and
   consumers no longer need to sort the components themselves. (#102)
 * Added `scripts/compile-cost` (`make compile-cost`), which measures what a
   consumer pays to compile against a checkout: the functions the compiler
   emits into a package that names each public type, how long such a package
   takes to recompile, and, with `-cold` or `-consumer`, a cold build's
   critical path or a real module's before-and-after timings. A guard test
   holds every probe under a threshold that comes down as the cost is cut.
   (#104)

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

// Package suitematerial loads one selected member's verified execution material
// from a local AgentSuite OCI layout or uncompressed OCI layout tar archive.
//
// Loading is read-only with respect to caller input. It returns exact immutable
// identities, copied instructions, and the complete portable model requirement.
// It does not select a runtime, composition, target, inference deployment, or
// credential; execute cases; verify evaluation evidence; publish; or Lift.
// Callers must not mutate a layout concurrently with LoadLayout.
package suitematerial

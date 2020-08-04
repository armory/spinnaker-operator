# v1.0.3
- refactor: transformer and (change) detectors will now be organized by functionality and kept in different modules.
- fix: a long time bug in `FreeForm` is also fixed. It was causing transformers that attempts to modify the config (or profiles) in memory to also leak the change into the operator's informer cache.
- fix: Validation webhook now patches the status. We cannot return the patches directly because we're changing the status. That should fix some validation errors trying to apply a new `SpinnakerService`
- fix: Validation service has ports named for Istio support
- fix: Crash when using `SpinnakerAccount` with sharded services ("HA mode")

# v1.0.0
TODO: need to backfill this

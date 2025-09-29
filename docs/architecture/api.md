# API `v0`

a rough outline of api areas

## identity

- user login to instance
- user logout of instance
- cycle public identity
  - affordance of cycling identity on specific networks

## messages

send messages from client to multiband, and let multiband manage the radios and how to send your message over various networks.

- send API
  - specific affordance to provide some path preferences (eg send to @someone prefer LXMF identity `abc` fallback to meshtastic identity `def`)
  - maybe also callbacks and read receipts, or some rudimentary retry/backoff logic for scripting fallthrough behavior)

## queue

outbound message flow control.

- queue operations on message cache and outbox (spool? need a non archaic name for these groups)
- show queue status
- delete messages from queue
  - bulk operations with filtering on useful message attributes

## files [v1]

file send/receive

supports: 

- ipfs
- magnet
- file upload
  - with read recipts/ul+dl metrics, and retry logic (from message routing layer)
  - per-file access controls
- receive by sha, magnet, other uri
  - support partial download across multiple interfaces (handle final merge of chunks for user)

## platform

- status of radios connected
- routing preferences
  - global routing logic (tbd how this complements per-message/per-identity routing presets)
- bring radios on and offline
- maybe serial based access to the underlying hw if supported?
- storage
  - manage storage usage
    - file uploads, partial downloads, message queues, etc


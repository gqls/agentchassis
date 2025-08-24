Key Features of This Implementation:

Configurable Timeout:

Default 5 minutes, but can be configured via child_timeout_minutes in step config
Different workflows can have different timeouts


Parent State Check:

Checks if parent is still waiting for this specific child
Only sends timeout if parent is actually waiting


Timeout Notification:

Sends a proper error response to parent via the response topic
Parent's HandleResponse will process this like any other response


Child Cleanup:

Optionally marks the child orchestration as failed if it's still running
Prevents zombie orchestrations


Proper Context Handling:

Uses a new context for the goroutine operations
Includes timeout for database operations


Comprehensive Logging:

Logs when timeout monitor starts
Logs timeout events
Logs when timeout check passes (child completed in time)



This implementation ensures that parent orchestrations don't wait forever for child orchestrations that may have failed silently or gotten stuck. The parent will receive a timeout notification and can continue or fail appropriately.
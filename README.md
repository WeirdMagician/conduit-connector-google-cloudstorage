# Conduit Connector Google Cloud Storage

### General
The Google Cloud Storage connector is one of [Conduit](https://github.com/ConduitIO/conduit) plugins. It provides a source GCS connectors.

### How to build it
Run `make`.

### Testing

Run `make test` to run all the tests. You must set the environment variables (`GCP_ServiceAccount_Key`
, `GCP_ProjectID`, `GCP_Bucket`)
before you run all the tests. If not set, the tests that use these variables will be ignored.

## GCS Source

The Google Cloud Storage Source Connector connects to a GCS bucket with the provided configurations, using serviceAccountKey,bucket details . Then will call `Configure` to parse the
configurations and make sure the bucket exists, If the bucket doesn't exist, or the permissions fail, then an error will
occur. After that, the
`Open` method is called to start the connection from the provided position.


### Change Data Capture (CDC)

This connector implements CDC features for GCS by scanning the bucket for changes every
`pollingPeriod` and detecting any change that happened after a certain timestamp. These changes (update, delete, insert)
are then inserted into a buffer that is checked on each Read request.

* To capture "delete" actions, the GCS bucket versioning must be enabled.
* To capture "insert" or "update" actions, the bucket versioning doesn't matter.


#### Position Handling

Here the position is constructed using the below custom type which includes the object key which was last read,Timestamp of the object concerned event, and type of the reading mode.

```
type Position struct {
	Key       string    `json:"key"`
	Timestamp time.Time `json:"timestamp"`
	Type      Type      `json:"type"`
}
```

The connector goes through two reading modes.

* Snapshot mode (Value 0): which loops through the GCS bucket and returns the objects that are already in there. The _position type_ during this mode is 0. which makes the connector know at what mode it is and what object it last
  read. The _position Timestamp_ will be used when changing to CDC mode, the iterator will capture changes that
  happened after that.

* CDC mode: (Value 1) this mode iterates through the GCS bucket every `pollingPeriod` and captures new actions made on the bucket.
  the _Position Type_ during this mode is 1. This position is used to return only the
  actions with a _Position Timestamp_ higher than the last record returned, which will ensure that no duplications are in
  place.


### Record Keys

The GCS object key uniquely identifies the objects in an Google Cloud Storage bucket, which is why a record key is the key read from
the GCS bucket.


### Configuration

The config passed to `Configure` can contain the following fields.

| name                  | description                                                                            | required  | example             |
|-----------------------|----------------------------------------------------------------------------------------|-----------|---------------------|
| `Service Account Key` | GCP service account key                                                                | yes       | "Service_Account_Key" |
| `GCS Bucket`          | the GCSbucket name                                                                 | yes       | "bucket_name"       |
| `pollingPeriod`       | polling period for the CDC mode, formatted as a time.Duration string. default is "1s"  | no        | "2s", "500ms"       |



### Known Limitations

* If a pipeline restarts during the snapshot, then the connector will start scanning the objects from the beginning of
  the bucket, which could result in duplications.

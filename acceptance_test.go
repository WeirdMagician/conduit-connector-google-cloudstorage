// Copyright © 2022 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connector

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"go.uber.org/goleak"

	"github.com/conduitio/conduit-connector-google-cloudstorage/config"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

type GCSAcceptanceTestDriver struct {
	sdk.ConfigurableAcceptanceTestDriver
	GCSClient *storage.Client
}

func (d GCSAcceptanceTestDriver) WriteToSource(t *testing.T, records []sdk.Record) []sdk.Record {
	ctx := context.Background()
	testBucket := d.SourceConfig(t)[config.ConfigKeyGCSBucket]
	for _, record := range records {
		wc := d.GCSClient.Bucket(testBucket).Object(string(record.Key.Bytes())).NewWriter(ctx)
		defer func() {
			if err := wc.Close(); err != nil {
				t.Fatal(err)
			}
		}()
		if _, err := wc.Write(record.Payload.Bytes()); err != nil {
			t.Fatal(err)
		}
	}
	return records
}

func TestAcceptance(t *testing.T) {
	sourceConfig, err := parseIntegrationConfig()
	if err != nil {
		t.Skip(err)
	}

	gcsClient, err := newGCSClient(sourceConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := gcsClient.Close(); err != nil {
			t.Fatal(err)
		}
	})

	sdk.AcceptanceTest(t, GCSAcceptanceTestDriver{
		sdk.ConfigurableAcceptanceTestDriver{
			Config: sdk.ConfigurableAcceptanceTestDriverConfig{
				Connector:         Connector,
				SourceConfig:      sourceConfig,
				DestinationConfig: nil,
				BeforeTest: func(t *testing.T) {
					sourceConfig[config.ConfigKeyGCSBucket] = "acceptance-test-bucket-" + uuid.NewString()
					if err := createTestGCSBucket(gcsClient, sourceConfig["projectID"], sourceConfig[config.ConfigKeyGCSBucket]); err != nil {
						t.Fatalf("could not create test gcs bucket: %v", err)
					}
				},
				AfterTest: func(t *testing.T) {
					clearAndDeleteTestGCSBucket(t, gcsClient, sourceConfig[config.ConfigKeyGCSBucket])
				},
				// Apart from the IgnoreCurrent, runtime_pollWait is also ignorned because here the GCS/storage client(Created Above) make a gRPC connection which is consistent and opens until it is closed.
				GoleakOptions: []goleak.Option{goleak.IgnoreCurrent(), goleak.IgnoreTopFunction("internal/poll.runtime_pollWait")},
				// Tests Starting from TestAcceptance/TestDestination were skipped because the destination connector for the GCS is not implemented.
				Skip: []string{"^TestAcceptance/TestDestination_"},
			},
		},
		gcsClient,
	})
}

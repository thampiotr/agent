package cloudwatch_exporter

import (
	"testing"

	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const configString = `
sts_region: us-east-2
discovery:
  exported_tags:
    AWS/EC2:
      - name
      - type
  jobs:
    - type: AWS/EC2
      search_tags:
        - key: instance_type
          value: spot
      regions:
        - us-east-2
      roles:
        - role_arn: arn:aws:iam::878167871295:role/yace_testing
      custom_tags:
        - key: alias
          value: tesis
      metrics:
        - name: CPUUtilization
          period: 5m
          statistics:
            - Maximum
            - Average
static:
  - regions:
      - us-east-2
    name: custom_tesis_metrics
    namespace: CoolApp
    dimensions:
      - name: PURCHASES_SERVICE
        value: CoolService
      - name: APP_VERSION
        value: 1.0
    metrics:
      - name: KPIs
        period: 5m
        statistics:
          - Average
`

// for testing fips_disabled behaviour
const configString2 = `
sts_region: us-east-2
fips_disabled: true
discovery:
  exported_tags:
    AWS/EC2:
      - name
      - type
  jobs:
    - type: AWS/EC2
      search_tags:
        - key: instance_type
          value: spot
      regions:
        - us-east-2
      roles:
        - role_arn: arn:aws:iam::878167871295:role/yace_testing
      custom_tags:
        - key: alias
          value: tesis
      metrics:
        - name: CPUUtilization
          period: 5m
          statistics:
            - Maximum
            - Average
static:
  - regions:
      - us-east-2
    name: custom_tesis_metrics
    namespace: CoolApp
    dimensions:
      - name: PURCHASES_SERVICE
        value: CoolService
      - name: APP_VERSION
        value: 1.0
    metrics:
      - name: KPIs
        period: 5m
        statistics:
          - Average
`

var falsePtr = false
var truePtr = true

var expectedConfig = yaceConf.ScrapeConf{
	APIVersion: "v1alpha1",
	StsRegion:  "us-east-2",
	Discovery: yaceConf.Discovery{
		ExportedTagsOnMetrics: map[string][]string{
			"AWS/EC2": {"name", "type"},
		},
		Jobs: []*yaceConf.Job{
			{
				Type:    "AWS/EC2",
				Regions: []string{"us-east-2"},
				Roles: []yaceConf.Role{
					{
						RoleArn: "arn:aws:iam::878167871295:role/yace_testing",
					},
				},
				CustomTags: []yaceModel.Tag{
					{
						Key:   "alias",
						Value: "tesis",
					},
				},
				SearchTags: []yaceModel.Tag{
					{
						Key:   "instance_type",
						Value: "spot",
					},
				},
				Metrics: []*yaceConf.Metric{
					{
						Name:       "CPUUtilization",
						Statistics: []string{"Maximum", "Average"},
						// Defaults get configured from general settings
						Period:                 300,
						Length:                 300,
						Delay:                  0,
						NilToZero:              &nilToZero,
						AddCloudwatchTimestamp: &addCloudwatchTimestamp,
					},
				},
				RoundingPeriod: nil,
				JobLevelMetricFields: yaceConf.JobLevelMetricFields{
					Period:                 0,
					Length:                 0,
					Delay:                  0,
					AddCloudwatchTimestamp: &falsePtr,
					NilToZero:              &nilToZero,
				},
			},
		},
	},
	Static: []*yaceConf.Static{
		{
			Name:       "custom_tesis_metrics",
			Regions:    []string{"us-east-2"},
			Roles:      []yaceConf.Role{{}},
			Namespace:  "CoolApp",
			CustomTags: []yaceModel.Tag{},
			Dimensions: []yaceConf.Dimension{
				{
					Name:  "PURCHASES_SERVICE",
					Value: "CoolService",
				},
				{
					Name:  "APP_VERSION",
					Value: "1.0",
				},
			},
			Metrics: []*yaceConf.Metric{
				{
					Name:                   "KPIs",
					Period:                 300,
					Length:                 300,
					Statistics:             []string{"Average"},
					Delay:                  0,
					NilToZero:              &nilToZero,
					AddCloudwatchTimestamp: &addCloudwatchTimestamp,
				},
			},
		},
	},
}

func TestTranslateConfigToYACEConfig(t *testing.T) {
	c := Config{}
	err := yaml.Unmarshal([]byte(configString), &c)
	require.NoError(t, err, "failed to unmarshall config")

	yaceConf, fipsEnabled, err := ToYACEConfig(&c)
	require.NoError(t, err, "failed to translate to YACE configuration")

	require.EqualValues(t, expectedConfig, yaceConf)
	require.EqualValues(t, truePtr, fipsEnabled)

	err = yaml.Unmarshal([]byte(configString2), &c)
	require.NoError(t, err, "failed to unmarshall config")

	yaceConf, fipsEnabled2, err := ToYACEConfig(&c)
	require.NoError(t, err, "failed to translate to YACE configuration")

	require.EqualValues(t, expectedConfig, yaceConf)
	require.EqualValues(t, falsePtr, fipsEnabled2)
}

func TestCloudwatchExporterConfigInstanceKey(t *testing.T) {
	cfg1 := &Config{
		STSRegion: "us-east-2",
	}
	cfg2 := &Config{
		STSRegion: "us-east-3",
	}

	cfg1Hash, err := cfg1.InstanceKey("")
	require.NoError(t, err)
	cfg2Hash, err := cfg2.InstanceKey("")
	require.NoError(t, err)

	assert.NotEqual(t, cfg1Hash, cfg2Hash)

	// test that making them equal in values leads to the same instance key
	cfg2.STSRegion = "us-east-2"
	cfg2Hash, err = cfg2.InstanceKey("")
	require.NoError(t, err)

	assert.Equal(t, cfg1Hash, cfg2Hash)
}

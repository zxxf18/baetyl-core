package shadow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/baetyl/baetyl-core/store"
	"github.com/stretchr/testify/assert"
)

func TestShadow(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	shad, err := NewShadow(t.Name(), t.Name(), s)
	assert.NoError(t, err)
	assert.NotNil(t, shad)

	// ! test sequence is important
	tests := []struct {
		name         string
		desired      string
		reported     string
		desireDelta  string
		reportDelta  string
		desireStored string
		reportStored string
		desireErr    error
		reportErr    error
	}{
		{
			name:         "1",
			desired:      "{}",
			reported:     "{}",
			desireDelta:  "{}",
			reportDelta:  "{}",
			desireStored: "{}",
			reportStored: "{}",
		},
		{
			name:         "2",
			desired:      `{"name": "module", "version": "45"}`,
			reported:     `{"name": "module", "version": "43"}`,
			desireDelta:  `{"name": "module", "version": "45"}`,
			reportDelta:  `{"version": "45"}`,
			desireStored: `{"name": "module", "version": "45"}`,
			reportStored: `{"name": "module", "version": "43"}`,
		},
		{
			name:         "3",
			desired:      `{"name": "module", "module": {"image": "test:v2"}}`,
			reported:     `{"name": "module", "module": {"image": "test:v1"}}`,
			desireDelta:  `{"version": "45", "module": {"image": "test:v2"}}`,
			reportDelta:  `{"version": "45", "module": {"image": "test:v2"}}`,
			desireStored: `{"name": "module", "version": "45", "module": {"image": "test:v2"}}`,
			reportStored: `{"name": "module", "version": "43", "module": {"image": "test:v1"}}`,
		},
		{
			name:         "4",
			desired:      `{"module": {"image": "test:v2", "array": []}}`,
			reported:     `{"module": {"image": "test:v1", "object": {"attr": "value"}}}`,
			desireDelta:  `{"version": "45", "module": {"image": "test:v2", "array": []}}`,
			reportDelta:  `{"version": "45", "module": {"image": "test:v2", "array": []}}`,
			desireStored: `{"name": "module", "version": "45", "module": {"image": "test:v2", "array": []}}`,
			reportStored: `{"name": "module", "version": "43", "module": {"image": "test:v1", "object": {"attr": "value"}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var desired, reported, desireDelta, reportDelta, desireStored, reportStored map[string]interface{}
			assert.NoError(t, json.Unmarshal([]byte(tt.desired), &desired))
			assert.NoError(t, json.Unmarshal([]byte(tt.reported), &reported))
			assert.NoError(t, json.Unmarshal([]byte(tt.desireDelta), &desireDelta))
			assert.NoError(t, json.Unmarshal([]byte(tt.reportDelta), &reportDelta))
			assert.NoError(t, json.Unmarshal([]byte(tt.desireStored), &desireStored))
			assert.NoError(t, json.Unmarshal([]byte(tt.reportStored), &reportStored))

			gotDelta, err := shad.Desire(desired)
			assert.Equal(t, tt.desireErr, err)
			if !reflect.DeepEqual(gotDelta, desireDelta) {
				t.Errorf("Shadow.Desire() = %v, want %v", gotDelta, desireDelta)
			}
			gotDelta, err = shad.Report(reported)
			assert.Equal(t, tt.reportErr, err)
			if !reflect.DeepEqual(gotDelta, reportDelta) {
				t.Errorf("Shadow.Report() = %v, want %v", gotDelta, reportDelta)
			}

			actual, err := shad.Get()
			assert.NoError(t, err)
			if actual.Desired == nil {
				assert.Empty(t, desireStored)
			} else {
				if !reflect.DeepEqual(actual.Desired, desireStored) {
					t.Errorf("Shadow.Get().Desired = %v, want %v", actual.Desired, desireStored)
				}
			}
			if actual.Reported == nil {
				assert.Empty(t, reportStored)
			} else {
				if !reflect.DeepEqual(actual.Reported, reportStored) {
					t.Errorf("Shadow.Get().Reported = %v, want %v", actual.Reported, reportStored)
				}
			}
		})
	}
}

func TestShadowRenew(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	shad, err := NewShadow(t.Name(), t.Name(), s)
	assert.NoError(t, err)
	assert.NotNil(t, shad)

	desire := map[string]interface{}{"apps": map[string]interface{}{"app1": "123", "app2": "234", "app3": "345", "app4": "456", "app5": ""}}
	delta, err := shad.Desire(desire)
	assert.NoError(t, err)
	apps := delta["apps"].(map[string]interface{})
	assert.Len(t, apps, 5)
	assert.Equal(t, "123", apps["app1"])
	assert.Equal(t, "234", apps["app2"])
	assert.Equal(t, "345", apps["app3"])
	assert.Equal(t, "456", apps["app4"])
	assert.Equal(t, "", apps["app5"])

	report := map[string]interface{}{"apps": map[string]interface{}{"app1": "123", "app2": "235", "app3": "", "app5": "567", "app6": "678"}}
	delta, err = shad.Report(report)
	assert.NoError(t, err)
	apps = delta["apps"].(map[string]interface{})
	assert.Len(t, apps, 4)
	assert.Equal(t, "234", apps["app2"])
	assert.Equal(t, "345", apps["app3"])
	assert.Equal(t, "456", apps["app4"])
	assert.Equal(t, "", apps["app5"])

	shad, err = NewShadow(t.Name(), t.Name(), s)
	assert.NoError(t, err)
	assert.NotNil(t, shad)

	delta, err = shad.Report(report)
	assert.NoError(t, err)
	apps = delta["apps"].(map[string]interface{})
	assert.Len(t, apps, 4)
	assert.Equal(t, "234", apps["app2"])
	assert.Equal(t, "345", apps["app3"])
	assert.Equal(t, "456", apps["app4"])
	assert.Equal(t, "", apps["app5"])
}
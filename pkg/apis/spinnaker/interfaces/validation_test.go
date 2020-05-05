package interfaces

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
	"time"
)

func TestNeedsValidation(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		seconds  int
		expected bool
	}{
		{
			name:     "first time needs validation",
			time:     time.Time{},
			seconds:  0,
			expected: true,
		},
		{
			name:     "when no frequency use default frequency",
			time:     time.Now(),
			seconds:  0,
			expected: false,
		},
		{
			name:     "long timeout no validation",
			time:     time.Now(),
			seconds:  99999,
			expected: false,
		},
		{
			name:     "past valid should validate again",
			time:     time.Now().Add(time.Duration(-20) * time.Second),
			seconds:  1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidationSetting{
				Enabled:          true,
				FrequencySeconds: intstr.FromInt(tt.seconds),
			}
			tm := metav1.NewTime(tt.time)
			assert.Equal(t, tt.expected, v.NeedsValidation(tm))
		})
	}
}

func TestSetAndGetHash(t *testing.T) {
	s := SpinnakerServiceStatus{}
	n := time.Now()
	p := s.UpdateHashIfNotExist("test", "abcdef", n, false)
	assert.Equal(t, "", p.Hash)
	assert.True(t, p.LastUpdatedAt.Time.IsZero())
	h := s.LastDeployed["test"]
	assert.Equal(t, "abcdef", h.Hash)
	assert.True(t, n.Equal(h.LastUpdatedAt.Time))

	// Now test with an existing hash
	n2 := time.Now()
	p = s.UpdateHashIfNotExist("test", "xyz", n2, false)
	assert.Equal(t, "abcdef", p.Hash)
	assert.True(t, n.Equal(p.LastUpdatedAt.Time))
	h = s.LastDeployed["test"]
	assert.Equal(t, "xyz", h.Hash)
	assert.False(t, n2.Equal(h.LastUpdatedAt.Time))

	// Now test while updating time
	n3 := time.Now()
	p = s.UpdateHashIfNotExist("test", "mnop", n3, true)
	assert.Equal(t, "xyz", p.Hash)
	assert.True(t, n.Equal(p.LastUpdatedAt.Time))
	h = s.LastDeployed["test"]
	assert.Equal(t, "mnop", h.Hash)
	assert.True(t, n3.Equal(h.LastUpdatedAt.Time))
}

package throttle

import (
	"testing"
	"time"
)

type call struct {
	nowOffset time.Duration
	count     int64
	exp       time.Duration
}

func Test_RateLimiter(t *testing.T) {
	tests := []struct {
		name     string
		capacity int64
		fill     time.Duration
		calls    []call
	}{
		{
			name:     "SeveralCalls",
			capacity: 1,
			fill:     time.Second,
			calls: []call{
				{
					nowOffset: 0,
					count:     0,
					exp:       0,
				},
				{
					nowOffset: time.Second,
					count:     1,
					exp:       time.Duration(0),
				},
				{
					nowOffset: time.Second,
					count:     1,
					exp:       time.Second,
				},
			},
		},
		{
			name:     "CapacityExeeded",
			capacity: 1,
			fill:     time.Second,
			calls: []call{
				{
					nowOffset: 0,
					count:     10,
					exp:       time.Second * 10,
				},
			},
		},
		{
			name:     "SubQuantTime",
			capacity: 5,
			fill:     time.Millisecond * 250,
			calls: []call{
				{
					nowOffset: 0,
					count:     1,
					exp:       time.Millisecond * 50,
				},
				{
					nowOffset: 0,
					count:     2,
					exp:       time.Millisecond * 150,
				},
				{
					nowOffset: 0,
					count:     8,
					exp:       time.Millisecond * 550,
				},
				{
					nowOffset: time.Millisecond * 300,
					count:     4,
					exp:       time.Millisecond * 450,
				},
			},
		},
	}

	for _, tc := range tests {
		lim := NewTBA(tc.fill, tc.capacity)
		t.Run(tc.name, func(t *testing.T) {
			for i, v := range tc.calls {
				got := lim.waitDuration(lim.initTime.Add(v.nowOffset), v.count)
				if got != v.exp {
					t.Errorf("Got: %v expected: %v in call[%v]", got, v.exp, i)
				}
			}
		})
	}
}

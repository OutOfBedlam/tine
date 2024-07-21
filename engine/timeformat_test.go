package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimformatter(t *testing.T) {
	tick := time.Unix(time.Now().Unix(), 0)

	nstf := NewTimeformatter("ns")
	require.True(t, nstf.IsEpoch())
	require.Equal(t, tick.UnixNano(), nstf.Epoch(tick))
	require.Equal(t, fmt.Sprintf("%d", tick.UnixNano()), nstf.Format(tick))
	_, err := nstf.Parse("not a number")
	require.Error(t, err)
	r, err := nstf.Parse(fmt.Sprintf("%d", tick.UnixNano()))
	require.NoError(t, err)
	require.Equal(t, tick.UnixNano(), r.UnixNano())

	ustf := NewTimeformatter("us")
	require.True(t, ustf.IsEpoch())
	require.Equal(t, tick.UnixNano()/1000, ustf.Epoch(tick))
	require.Equal(t, fmt.Sprintf("%d", tick.UnixNano()/1000), ustf.Format(tick))
	_, err = ustf.Parse("not a number")
	require.Error(t, err)
	r, err = ustf.Parse(fmt.Sprintf("%d", tick.UnixNano()/1000))
	require.NoError(t, err)
	require.Equal(t, tick.UnixNano(), r.UnixNano())

	mstf := NewTimeformatter("ms")
	require.True(t, mstf.IsEpoch())
	require.Equal(t, tick.UnixNano()/1000_000, mstf.Epoch(tick))
	require.Equal(t, fmt.Sprintf("%d", tick.UnixNano()/1000_000), mstf.Format(tick))
	_, err = mstf.Parse("not a number")
	require.Error(t, err)
	r, err = mstf.Parse(fmt.Sprintf("%d", tick.UnixNano()/1000_000))
	require.NoError(t, err)
	require.Equal(t, tick.UnixNano(), r.UnixNano())

	stf := NewTimeformatter("s")
	require.True(t, stf.IsEpoch())
	require.Equal(t, tick.UnixNano()/1000_000_000, stf.Epoch(tick))
	_, err = stf.Parse("not a number")
	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("%d", tick.UnixNano()/1000_000_000), stf.Format(tick))
	r, err = stf.Parse(fmt.Sprintf("%d", tick.UnixNano()/1000_000_000))
	require.NoError(t, err)
	require.Equal(t, tick.UnixNano(), r.UnixNano())

	tz, err := time.LoadLocation("Asia/Seoul")
	require.NoError(t, err)
	tf := NewTimeformatterWithLocation("2006-01-02 15:04:05", tz)
	parsed, err := tf.Parse("2024-07-21 15:09:29")
	require.NoError(t, err)
	require.Equal(t, int64(1721542169), parsed.Unix())
}

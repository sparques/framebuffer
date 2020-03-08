// This file is subject to a 1-clause BSD license.
// Its contents can be found in the enclosed LICENSE file.

package framebuffer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

var (
	endmode      = []byte("endmode")
	regLabel     = regexp.MustCompile(`^mode\W+"([^"]+)"`)
	regGeometry  = regexp.MustCompile(`geometry\W+(\d+)\W+(\d+)\W+(\d+)\W+(\d+)\W+(\d+)`)
	regTimings   = regexp.MustCompile(`timings\W+(\d+)\W+(\d+)\W+(\d+)\W+(\d+)\W+(\d+)\W+(\d+)\W+(\d+)`)
	regHsync     = regexp.MustCompile(`hsync\W+high`)
	regVsync     = regexp.MustCompile(`vsync\W+high`)
	regCsync     = regexp.MustCompile(`csync\W+high`)
	regGsync     = regexp.MustCompile(`gsync\W+high`)
	regAccel     = regexp.MustCompile(`accel\W+true`)
	regBcast     = regexp.MustCompile(`bcast\W+true`)
	regGrayscale = regexp.MustCompile(`grayscale\W+true`)
	regExtsync   = regexp.MustCompile(`extsync\W+true`)
	regNonstd    = regexp.MustCompile(`nonstd\W+(\d+)`)
	regLaced     = regexp.MustCompile(`laced\W+true`)
	regDouble    = regexp.MustCompile(`double\W+true`)
	regFormat    = regexp.MustCompile(`rgba\W+(\d+)/(\d+),(\d+)/(\d+),(\d+)/(\d+),(\d+)/(\d+)`)
)

func readInt(v []byte, bits int) int {
	n, err := strconv.ParseInt(string(v), 10, bits)
	if err != nil {
		panic(err)
	}
	return int(n)
}

// readFBModes reads display mode data from the given stream.
// This is expected to come in the format defined at
// http://manned.org/fb.modes/81e6dc49
func readFBModes(r io.Reader) (list []*DisplayMode, err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("%v", x)
		}
	}()

	var line []byte

	rdr := bufio.NewReader(r)
	dm := new(DisplayMode)

	for {
		line, err = rdr.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// End of mode?
		if bytes.Contains(line, endmode) {
			list = append(list, dm)
			dm = new(DisplayMode)
			continue
		}

		// Parse name
		matches := regLabel.FindSubmatch(line)
		if len(matches) > 1 {
			dm.Name = string(matches[1])
			continue
		}

		// Parse nonstd
		matches = regNonstd.FindSubmatch(line)
		if len(matches) > 1 {
			dm.Nonstandard = readInt(matches[1], 32)
			continue
		}

		// Parse hsync
		if regHsync.Match(line) {
			dm.Sync |= SyncHorHighAct
			continue
		}

		// Parse vsync
		if regVsync.Match(line) {
			dm.Sync |= SyncVertHighAct
			continue
		}

		// Parse csync
		if regCsync.Match(line) {
			dm.Sync |= SyncCompHighAct
			continue
		}

		// Parse gsync
		if regGsync.Match(line) {
			dm.Sync |= SyncOnGreen
			continue
		}

		// Parse bcast
		if regBcast.Match(line) {
			dm.Sync |= SyncBroadcast
			continue
		}

		// Parse extsync
		if regExtsync.Match(line) {
			dm.Sync |= SyncExt
			continue
		}

		// Parse accel
		if regAccel.Match(line) {
			dm.Accelerated = true
			continue
		}

		// Parse grayscale
		if regGrayscale.Match(line) {
			dm.Grayscale = true
			continue
		}

		// Parse laced
		if regLaced.Match(line) {
			dm.VMode |= VModeInterlaced
			continue
		}

		// Parse double
		if regDouble.Match(line) {
			dm.VMode |= VModeDouble
			continue
		}

		// Parse geometry
		matches = regGeometry.FindSubmatch(line)
		if len(matches) > 1 {
			dm.Geometry.XRes = readInt(matches[1], 32)
			dm.Geometry.YRes = readInt(matches[2], 32)
			dm.Geometry.XVRes = readInt(matches[3], 32)
			dm.Geometry.YVRes = readInt(matches[4], 32)
			dm.Geometry.Depth = readInt(matches[5], 32)
		}

		// Parse timings
		matches = regTimings.FindSubmatch(line)
		if len(matches) > 1 {
			dm.Timings.Pixclock = readInt(matches[1], 32)
			dm.Timings.Left = readInt(matches[2], 32)
			dm.Timings.Right = readInt(matches[3], 32)
			dm.Timings.Upper = readInt(matches[4], 32)
			dm.Timings.Lower = readInt(matches[5], 32)
			dm.Timings.HSLen = readInt(matches[6], 32)
			dm.Timings.VSLen = readInt(matches[7], 32)
		}

		// Parse pixel format
		matches = regFormat.FindSubmatch(line)
		if len(matches) > 1 {
			dm.Format.RedBits = uint8(readInt(matches[1], 8))
			dm.Format.RedShift = uint8(readInt(matches[1], 8))
			dm.Format.GreenBits = uint8(readInt(matches[1], 8))
			dm.Format.GreenShift = uint8(readInt(matches[1], 8))
			dm.Format.BlueBits = uint8(readInt(matches[1], 8))
			dm.Format.BlueShift = uint8(readInt(matches[1], 8))
			dm.Format.AlphaBits = uint8(readInt(matches[1], 8))
			dm.Format.AlphaShift = uint8(readInt(matches[1], 8))
		}
	}

	return
}

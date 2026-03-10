package dashboard

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/harmonica"
	"github.com/devkeshwani/termf1/internal/api/liveserver"
	"github.com/devkeshwani/termf1/internal/api/openf1"
)

// buildLiveRows converts a live State snapshot into sorted liveRow slice,
// reusing harmonica spring state from the previous frame for smooth gap animation.
func buildLiveRows(st *liveserver.State, prev []liveRow) []liveRow {
	if st == nil || len(st.Timing) == 0 {
		return nil
	}

	prevMap := make(map[string]*liveRow, len(prev))
	for i := range prev {
		prevMap[prev[i].Number] = &prev[i]
	}

	rows := make([]liveRow, 0, len(st.Timing))
	for num, t := range st.Timing {
		di := st.Drivers[num]
		tla := num
		teamName := ""
		teamColour := ""
		if di != nil {
			tla = di.Tla
			teamName = di.TeamName
			teamColour = di.TeamColour
		}

		s1, s2, s3 := "", "", ""
		s1Prev, s2Prev, s3Prev := "", "", ""
		s1Fast, s2Fast, s3Fast := false, false, false
		s1Personal, s2Personal, s3Personal := false, false, false
		var s1Segs, s2Segs, s3Segs []int
		if len(t.Sectors) >= 1 {
			s1 = t.Sectors[0].Value
			s1Prev = t.Sectors[0].PreviousValue
			s1Fast = t.Sectors[0].OverallFastest
			s1Personal = t.Sectors[0].PersonalFastest
			for _, sg := range t.Sectors[0].Segments {
				s1Segs = append(s1Segs, sg.Status)
			}
		}
		if len(t.Sectors) >= 2 {
			s2 = t.Sectors[1].Value
			s2Prev = t.Sectors[1].PreviousValue
			s2Fast = t.Sectors[1].OverallFastest
			s2Personal = t.Sectors[1].PersonalFastest
			for _, sg := range t.Sectors[1].Segments {
				s2Segs = append(s2Segs, sg.Status)
			}
		}
		if len(t.Sectors) >= 3 {
			s3 = t.Sectors[2].Value
			s3Prev = t.Sectors[2].PreviousValue
			s3Fast = t.Sectors[2].OverallFastest
			s3Personal = t.Sectors[2].PersonalFastest
			for _, sg := range t.Sectors[2].Segments {
				s3Segs = append(s3Segs, sg.Status)
			}
		}

		// Tyre data from TimingAppData.
		compound, tyreAge, tyreNew := "", 0, false
		if st.Stints != nil {
			if stintData, ok := st.Stints[num]; ok && stintData != nil {
				compound = stintData.Compound
				tyreAge = stintData.TotalLaps
				tyreNew = stintData.New
			}
		}

		// DRS status from CarData channel 45: 10/12/14 = open.
		drs := 0
		if st.CarData != nil {
			if cd, ok := st.CarData[num]; ok && cd != nil {
				drs = cd.Channels.DRS
			}
		}

		r := liveRow{
			Pos:              t.Line,
			Number:           num,
			Tla:              tla,
			TeamName:         teamName,
			TeamColour:       teamColour,
			GapToLeader:      t.GapToLeader,
			Interval:         t.IntervalToPositionAhead.Value,
			IntervalCatching: t.IntervalToPositionAhead.Catching,
			LastLap:          t.LastLapTime.Value,
			LastFastest:      t.LastLapTime.OverallFastest,
			LastPersonal:     t.LastLapTime.PersonalFastest,
			BestLap:          t.BestLapTime.Value,
			S1: s1, S2: s2, S3: s3,
			S1Prev: s1Prev, S2Prev: s2Prev, S3Prev: s3Prev,
			S1Fast: s1Fast, S2Fast: s2Fast, S3Fast: s3Fast,
			S1Personal: s1Personal, S2Personal: s2Personal, S3Personal: s3Personal,
			S1Segs: s1Segs, S2Segs: s2Segs, S3Segs: s3Segs,
			SpeedST:  t.Speeds.ST.Value,
			InPit:    t.InPit,
			PitCount: t.NumberOfPitStops,
			Laps:     t.NumberOfLaps,
			Retired:  t.Retired,
			Compound: compound,
			TyreAge:  tyreAge,
			TyreNew:  tyreNew,
			DRS:      drs,
		}

		if p, ok := prevMap[num]; ok {
			r.gapSpring = p.gapSpring
			r.gapDisplay = p.gapDisplay
			// Carry previous position so we can show ▲/▼ in the UI.
			if p.Pos != r.Pos && p.Pos > 0 {
				r.PrevPos = p.Pos
			}
		} else {
			r.gapSpring = harmonica.NewSpring(harmonica.FPS(60), 6, 0.6)
		}
		if gapF, err := parseGapFloat(t.GapToLeader); err == nil {
			r.gapTarget = gapF
			r.gapStr = ""
		} else {
			r.gapStr = t.GapToLeader
		}

		rows = append(rows, r)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Pos == 0 {
			return false
		}
		if rows[j].Pos == 0 {
			return true
		}
		return rows[i].Pos < rows[j].Pos
	})
	return rows
}

// parseGapFloat strips a leading "+" and parses the gap value as float64.
func parseGapFloat(s string) (float64, error) {
	s = strings.TrimPrefix(s, "+")
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// buildFallbackRows adapts OpenF1 DriverRow data into fallbackRow slice.
func buildFallbackRows(p *openf1.DashboardPayload) []fallbackRow {
	if p == nil {
		return nil
	}
	rows := make([]fallbackRow, 0)
	for _, row := range buildRows(p) {
		rows = append(rows, fallbackRow{
			Pos:         row.Position,
			Number:      row.Number,
			Acronym:     row.Acronym,
			TeamName:    row.TeamName,
			TeamColor:   row.TeamColor,
			GapToLeader: row.GapToLeader,
			Interval:    row.Interval,
			LastLap:     row.LastLap,
			BestLap:     row.BestLap,
			Sector1:     row.Sector1,
			Sector2:     row.Sector2,
			Sector3:     row.Sector3,
			SpeedTrap:   row.SpeedTrap,
			Compound:    row.Compound,
			TyreAge:     row.TyreAge,
			PitCount:    row.PitCount,
			LapNumber:   row.LapNumber,
			DNF:         row.DNF,
		})
	}
	return rows
}

// liveRCMessages returns the last n race control messages from a live State.
func liveRCMessages(st *liveserver.State, n int) []string {
	if st == nil {
		return nil
	}
	out := make([]string, 0, len(st.RaceControlMessages))
	for _, m := range st.RaceControlMessages {
		out = append(out, m.Message)
	}
	if len(out) > n {
		out = out[len(out)-n:]
	}
	return out
}

// fallbackRCMessages returns the last n race control messages from OpenF1 data.
func fallbackRCMessages(p *openf1.DashboardPayload, n int) []string {
	if p == nil {
		return nil
	}
	return lastRCMessages(p.RaceControls, n)
}

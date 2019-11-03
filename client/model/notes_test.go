package model

import (
	"fmt"
	"testing"

	_ "alda.io/client/testing"
)

func expectNoteOffsets(expectedOffsets ...OffsetMs) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedOffsets) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedOffsets),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedOffsets); i++ {
			expectedOffset := expectedOffsets[i]
			actualOffset := s.Events[i].(NoteEvent).Offset
			if !equalish(expectedOffset, actualOffset) {
				return fmt.Errorf(
					"expected note #%d to have offset %f, but it was %f",
					i+1,
					expectedOffset,
					actualOffset,
				)
			}
		}

		return nil
	}
}

func expectNoteFloatValues(
	valueName string, method func(NoteEvent) float32, expectedValues []float32,
) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedValues) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedValues),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedValues); i++ {
			expectedValue := expectedValues[i]
			actualValue := method(s.Events[i].(NoteEvent))
			if !equalish32(expectedValue, actualValue) {
				return fmt.Errorf(
					"expected note #%d to have %s %f, but it was %f",
					i+1, valueName, expectedValue, actualValue,
				)
			}
		}

		return nil
	}
}

func expectNoteDurations(expectedDurations ...float32) func(*Score) error {
	return expectNoteFloatValues(
		"audible duration",
		func(note NoteEvent) float32 { return note.Duration },
		expectedDurations,
	)
}

func expectNoteAudibleDurations(
	expectedAudibleDurations ...float32,
) func(*Score) error {
	return expectNoteFloatValues(
		"audible duration",
		func(note NoteEvent) float32 { return note.AudibleDuration },
		expectedAudibleDurations,
	)
}

func expectMidiNoteNumbers(expectedNoteNumbers ...int32) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedNoteNumbers) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedNoteNumbers),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedNoteNumbers); i++ {
			expectedNoteNumber := expectedNoteNumbers[i]
			actualNoteNumber := s.Events[i].(NoteEvent).MidiNote
			if expectedNoteNumber != actualNoteNumber {
				return fmt.Errorf(
					"expected note #%d to be MIDI note %d, but it was MIDI note %d",
					i+1,
					expectedNoteNumber,
					actualNoteNumber,
				)
			}
		}

		return nil
	}
}

func TestNotes(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "notes with provided durations",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
				Note{
					NoteLetter: D,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 8},
						},
					},
				},
				Note{
					NoteLetter: E,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 2222},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 1500, 1750),
				expectNoteDurations(1500, 250, 2222),
				expectMidiNoteNumbers(60, 62, 64),
			},
		},
		scoreUpdateTestCase{
			label: "implicit note duration",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
				Note{NoteLetter: D},
				Note{
					NoteLetter:  D,
					Accidentals: []Accidental{Sharp},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 50},
						},
					},
				},
				Note{NoteLetter: E},
				Note{
					NoteLetter: F,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 8},
						},
					},
				},
				Note{NoteLetter: G},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 1500, 3000, 3050, 3100, 3350),
				expectNoteDurations(1500, 1500, 50, 50, 250, 250),
				expectMidiNoteNumbers(60, 62, 63, 64, 65, 67),
			},
		},
		scoreUpdateTestCase{
			label: "note with 100% quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 100},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(500),
			},
		},
		scoreUpdateTestCase{
			label: "note with 90% quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 90},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(450),
			},
		},
		scoreUpdateTestCase{
			label: "note with 0% quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 0},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(),
			},
		},
		scoreUpdateTestCase{
			label: "slurred notes ignore quantization #1",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 90},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
					Slurred: true,
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(500),
			},
		},
		scoreUpdateTestCase{
			label: "slurred notes ignore quantization #2",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 90},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
					Slurred: true,
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(1000),
			},
		},
	)
}
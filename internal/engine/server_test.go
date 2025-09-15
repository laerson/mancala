package engine

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	enginepb "github.com/laerson/mancala/proto/engine"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestMove(t *testing.T) {
	type tc struct {
		name         string
		req          *enginepb.MoveRequest
		wantResponse *enginepb.MoveResponse
	}

	tests := []tc{
		{
			name: "sow and skip opponent store (P1 from pit 5 with 8 seeds)",
			req: &enginepb.MoveRequest{
				GameState: &enginepb.GameState{
					Board: &enginepb.Board{
						Pits: []uint32{4, 4, 4, 4, 4, 8, 0, 1, 1, 1, 1, 1, 1, 0},
					},
					CurrentPlayer: enginepb.Player_PLAYER_ONE,
				},
				PitIndex: 5,
			},
			wantResponse: &enginepb.MoveResponse{
				Result: &enginepb.MoveResponse_MoveResult{
					MoveResult: &enginepb.MoveResult{
						Board: &enginepb.Board{
							Pits: []uint32{5, 4, 4, 4, 4, 0, 1, 2, 2, 2, 2, 2, 2, 0},
						},
						CurrentPlayer: enginepb.Player_PLAYER_TWO,
						IsFinished:    false,
						Winner:        enginepb.Winner_NO_WINNER,
					},
				},
			},
		},
		{
			name: "capture for player one",
			req: &enginepb.MoveRequest{
				GameState: &enginepb.GameState{
					Board: &enginepb.Board{
						Pits: []uint32{1, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0},
					},
					CurrentPlayer: enginepb.Player_PLAYER_ONE,
				},
				PitIndex: 0,
			},
			wantResponse: &enginepb.MoveResponse{
				Result: &enginepb.MoveResponse_MoveResult{
					MoveResult: &enginepb.MoveResult{
						Board: &enginepb.Board{
							Pits: []uint32{0, 0, 0, 0, 0, 1, 2, 1, 1, 1, 1, 0, 1, 0},
						},
						CurrentPlayer: enginepb.Player_PLAYER_TWO,
						IsFinished:    false,
						Winner:        enginepb.Winner_NO_WINNER,
					},
				},
			},
		},
		{
			name: "extra turn",
			req: &enginepb.MoveRequest{
				GameState: &enginepb.GameState{
					Board: &enginepb.Board{
						Pits: []uint32{1, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0},
					},
					CurrentPlayer: enginepb.Player_PLAYER_ONE,
				},
				PitIndex: 5,
			},
			wantResponse: &enginepb.MoveResponse{
				Result: &enginepb.MoveResponse_MoveResult{
					MoveResult: &enginepb.MoveResult{
						Board: &enginepb.Board{
							Pits: []uint32{1, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0},
						},
						CurrentPlayer: enginepb.Player_PLAYER_ONE,
						IsFinished:    false,
						Winner:        enginepb.Winner_NO_WINNER,
					},
				},
			},
		},
		{
			name: "Finish game",
			req: &enginepb.MoveRequest{
				GameState: &enginepb.GameState{
					Board: &enginepb.Board{
						Pits: []uint32{0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0},
					},
					CurrentPlayer: enginepb.Player_PLAYER_ONE,
				},
				PitIndex: 5,
			},
			wantResponse: &enginepb.MoveResponse{
				Result: &enginepb.MoveResponse_MoveResult{
					MoveResult: &enginepb.MoveResult{
						Board: &enginepb.Board{
							Pits: []uint32{0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 6},
						},
						IsFinished: true,
						Winner:     enginepb.Winner_WINNER_PLAYER_TWO,
					},
				},
			},
		},
		{
			name: "Invalid Board",
			req: &enginepb.MoveRequest{
				GameState: &enginepb.GameState{
					Board: &enginepb.Board{
						Pits: []uint32{0, 0, 0, 0, 0},
					},
					CurrentPlayer: enginepb.Player_PLAYER_ONE,
				},
				PitIndex: 5,
			},
			wantResponse: &enginepb.MoveResponse{
				Result: &enginepb.MoveResponse_Error{
					Error: &enginepb.Error{
						Message: errInvalidBoard.Error(),
					},
				},
			},
		},
		{
			name: "Invalid Pit Index",
			req: &enginepb.MoveRequest{
				GameState: &enginepb.GameState{
					Board: &enginepb.Board{
						Pits: []uint32{4, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0},
					},
					CurrentPlayer: enginepb.Player_PLAYER_ONE,
				},
				PitIndex: 14,
			},
			wantResponse: &enginepb.MoveResponse{
				Result: &enginepb.MoveResponse_Error{
					Error: &enginepb.Error{
						Message: errInvalidPitIndex.Error(),
					},
				},
			},
		},
		{
			name: "Empty Pit",
			req: &enginepb.MoveRequest{
				GameState: &enginepb.GameState{
					Board: &enginepb.Board{
						Pits: []uint32{4, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0},
					},
					CurrentPlayer: enginepb.Player_PLAYER_ONE,
				},
				PitIndex: 5,
			},
			wantResponse: &enginepb.MoveResponse{
				Result: &enginepb.MoveResponse_Error{
					Error: &enginepb.Error{
						Message: errEmptyPit.Error(),
					},
				},
			},
		},
	}

	s := &Server{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResponse, err := s.Move(context.Background(), tt.req)
			if err != nil {
				t.Errorf("Move() error = %v", err)
			}
			if !cmp.Equal(gotResponse, tt.wantResponse, protocmp.Transform()) {
				t.Errorf("Move() gotResponse = %v, want %v", gotResponse, tt.wantResponse)
			}
		})
	}

}

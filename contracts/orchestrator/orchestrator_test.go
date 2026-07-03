package orchestrator

import (
	"testing"
)

func TestMissionStatus_Progress(t *testing.T) {
	tests := []struct {
		name       string
		totalTasks int
		doneTasks  int
		want       int
	}{
		{
			name:       "zero tasks",
			totalTasks: 0,
			doneTasks:  0,
			want:       0,
		},
		{
			name:       "half completed",
			totalTasks: 10,
			doneTasks:  5,
			want:       50,
		},
		{
			name:       "all completed",
			totalTasks: 3,
			doneTasks:  3,
			want:       100,
		},
		{
			name:       "no completed",
			totalTasks: 5,
			doneTasks:  0,
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MissionStatus{
				TotalTasks: tt.totalTasks,
				DoneTasks:  tt.doneTasks,
			}
			if got := s.Progress(); got != tt.want {
				t.Errorf("MissionStatus.Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}

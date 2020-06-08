package main

import (
	"container/heap"
	"context"
	"log"
	"time"
)

// ScheduleMovie stores the informations to
// schedule a movie.
type ScheduledMovie struct {
	Time     time.Time
	MovieID  int
	UserName string
	Email    string
}

// ScheduleList is list of movies to remeber
// user to watch at determined time.
// ScheduleList works as a min heap.
type ScheduleList []*ScheduledMovie

// Define the functions necessary to use heap.

func (sl ScheduleList) Len() int { return len(sl) }

func (sl ScheduleList) Less(i, j int) bool {
	t1, t2 := sl[i].Time, sl[j].Time
	return t1.Sub(t2) < 0
}

func (sl ScheduleList) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}

func (sl *ScheduleList) Push(x interface{}) {
	movie := x.(*ScheduledMovie)
	*sl = append(*sl, movie)
}

func (sl *ScheduleList) Pop() interface{} {
	old := *sl
	n := len(old)
	movie := old[n-1]
	old[n-1] = nil
	*sl = old[:n-1]
	return movie
}

// popBeforeTime remove scheduled movies that Time is berfore t.
func (sl *ScheduleList) popBeforeTime(t time.Time) []*ScheduledMovie {
	out := make([]*ScheduledMovie, 0)
	// While heap is not empty try to remove.
	for sl.Len() > 0 {
		// Remove a item from heap and verify if
		// Time is before t.
		register := heap.Pop(sl).(*ScheduledMovie)
		if register.Time.Before(t) {
			out = append(out, register)
		} else {
			// If Time is not before t push the
			// item to heap and stop removing,
			// the next item.Time is greater than
			// this one.
			heap.Push(sl, register)
			break
		}
	}
	return out
}

// schedule checks, periodically, if server should send
// email to users to rember of some movie.
func (s *server) schedule(ctx context.Context) {
	// Start the heap.
	s.mu.Lock()
	heap.Init(s.scheduleList)
	s.mu.Unlock()
	ticker := time.NewTicker(time.Minute * 1)
	for {
		select {
		// Cancelation. Stop function.
		case <-ctx.Done():
			return
		// It's time to check.
		case <-ticker.C:
			// Check if there are emails to send.
			now := time.Now()
			s.mu.Lock()
			registers := s.scheduleList.popBeforeTime(now)
			s.mu.Unlock()
			// If there are emails, send it.
			for _, r := range registers {
				// Send a email to user Email and MovieID, concurrently.
				go func(r *ScheduledMovie) {
					err := s.mailer.SendScheduledMovie(r.UserName, r.Email, r.MovieID)
					if err != nil {
						// If error just log. Best effort.
						log.Println(err)
					}
				}(r)
			}
		}
	}
}

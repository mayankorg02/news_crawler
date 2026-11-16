package internal

import (
	"context"
	"errors"
	"sync"
)

type Pipeline[In, Out any] struct {
	workers int
	process func(context.Context, In) Out
}

func NewPipeline[In, Out any](workers int, process func(context.Context, In) Out) (*Pipeline[In, Out], error) {
	if workers < 1 {
		return nil, errors.New("pipeline requires at least one worker")
	}
	if process == nil {
		return nil, errors.New("pipeline requires a process function")
	}

	return &Pipeline[In, Out]{
		workers: workers,
		process: process,
	}, nil
}

func (p *Pipeline[In, Out]) Run(ctx context.Context, input []In) []Out {
	jobs := make(chan In)
	results := make(chan Out)

	var workers sync.WaitGroup
	workers.Add(p.workers)
	for range p.workers {
		go func() {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-jobs:
					if !ok {
						return
					}

					result := p.process(ctx, item)
					select {
					case results <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, item := range input {
			select {
			case jobs <- item:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	var output []Out
	for result := range results {
		output = append(output, result)
	}
	return output
}

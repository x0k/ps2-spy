package retryable

import (
	"context"
	"time"
)

// Thread safe
type Retryable struct {
	try              func(ctx context.Context) error
	SuspenseDuration time.Duration
	Err              error
}

func New(action func(ctx context.Context) error) *Retryable {
	return &Retryable{
		try:              action,
		SuspenseDuration: 1 * time.Second,
	}
}

func (r Retryable) Run(ctx context.Context, options ...any) error {
	conditions := make([]func(*Retryable) bool, 0, len(options))
	beforeSuspense := make([]func(*Retryable), 0, len(options))
	for _, option := range options {
		switch v := option.(type) {
		case func(*Retryable) bool:
			conditions = append(conditions, v)
		case func(*Retryable):
			beforeSuspense = append(beforeSuspense, v)
		}
	}
	t := time.NewTimer(0)
	defer t.Stop()
	if !t.Stop() {
		<-t.C
	}
	for {
		r.Err = r.try(ctx)
		for _, condition := range conditions {
			if !condition(&r) {
				return r.Err
			}
		}
		for _, beforeSuspense := range beforeSuspense {
			beforeSuspense(&r)
		}
		t.Reset(r.SuspenseDuration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
		r.SuspenseDuration *= 2
	}
}

// Thread unsafe
type WithReturn[R any] struct {
	ret *Retryable
	res R
	err error
}

func NewWithReturn[R any](
	action func(ctx context.Context) (R, error),
) *WithReturn[R] {
	r2 := &WithReturn[R]{}
	r2.ret = New(func(ctx context.Context) error {
		r2.res, r2.err = action(ctx)
		return r2.err
	})
	return r2
}

func (r *WithReturn[R]) Run(ctx context.Context, options ...any) (R, error) {
	r.ret.Run(ctx, options...)
	return r.res, r.err
}

// Thread unsafe
type WithArg[A any, R any] struct {
	ret *Retryable
	arg A
	res R
	err error
}

func NewWithArg[A any, R any](
	action func(context.Context, A) (R, error),
) *WithArg[A, R] {
	r3 := &WithArg[A, R]{}
	r3.ret = New(func(ctx context.Context) error {
		r3.res, r3.err = action(ctx, r3.arg)
		return r3.err
	})
	return r3
}

func (r *WithArg[A, R]) Run(ctx context.Context, arg A, options ...any) (R, error) {
	r.arg = arg
	r.ret.Run(ctx, options...)
	return r.res, r.err
}

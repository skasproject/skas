package protector

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"sync"
	"time"
)

// The Protector is a mechanism of protection against BFA.
// It introduce an increasing delay on response in case of failure on a given login
// After some period without failure, the history is cleanup up.
//

type Protector interface {
	Entry(login string) (id int64, locked bool)
	Success(id int64, login string)
	Failure(id int64, login string)
	Exit(id int64, login string)
}

var _ Protector = &protector{}

type loginState struct {
	lastFailure  time.Time
	nbrOfFailure int64
}

type protector struct {
	mu               sync.Mutex
	id               int64
	logger           logr.Logger
	stateByLogin     map[string]*loginState
	cleanerPeriod    time.Duration
	cleanDelay       time.Duration
	freeFailure      int64 // No delay introduced up to this value
	maxPenalty       time.Duration
	penaltyByFailure time.Duration
}

type Option func(*protector)

// WithCleanerPeriod define the period if the cleanup processing
func WithCleanerPeriod(cleanerPeriod time.Duration) Option {
	return func(p *protector) {
		p.cleanerPeriod = cleanerPeriod
	}
}

// WithCleanDelay For a login, the failure history is cleaned up if there no new failure during this delay
func WithCleanDelay(cd time.Duration) Option {
	return func(p *protector) {
		p.cleanDelay = cd
	}
}

// WithFreeFailure Nbr of failure allowed before introducing a delay
func WithFreeFailure(ff int64) Option {
	return func(p *protector) {
		p.freeFailure = ff
	}
}

// WithMaxPenalty The introduced delay is capped to this value
func WithMaxPenalty(mp time.Duration) Option {
	return func(p *protector) {
		p.maxPenalty = mp
	}
}

// WithPenaltyByFailure Increment step introduced by failure
func WithPenaltyByFailure(pbf time.Duration) Option {
	return func(p *protector) {
		p.penaltyByFailure = pbf
	}
}

func New(ctx context.Context, logger logr.Logger, opts ...Option) Protector {
	p := &protector{
		logger:           logger,
		stateByLogin:     make(map[string]*loginState),
		cleanerPeriod:    60 * time.Second,
		cleanDelay:       30 * time.Minute,
		freeFailure:      2,
		maxPenalty:       15 * time.Second,
		penaltyByFailure: 1 * time.Second,
	}
	for _, opt := range opts {
		opt(p)
	}
	logger.Info("Cleaner start")
	go wait.Until(func() {
		p.clean()
	}, p.cleanerPeriod, ctx.Done())
	return p
}

func (p *protector) Entry(login string) (id int64, locked bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.id++
	p.logger.V(2).Info("protector.Entry()", "login", login, "id", id)
	return id, false
}

func (p *protector) Success(id int64, login string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logger.V(2).Info("protector.Success()", "login", login, "id", id)
}

func (p *protector) Failure(id int64, login string) {
	p.logger.V(2).Info("protector.Failure(1/2)", "login", login, "id", id)
	p.mu.Lock()
	state, ok := p.stateByLogin[login]
	if !ok {
		state = &loginState{}
		p.stateByLogin[login] = state
	}
	state.lastFailure = time.Now()
	state.nbrOfFailure++
	nbrOfFailure := state.nbrOfFailure
	p.mu.Unlock()
	delay := p.delayFromFailureCount(nbrOfFailure)
	p.logger.V(0).Info("protector.failure", "login", login, "failureCount", nbrOfFailure, "delay", delay.String())
	time.Sleep(delay)
	p.logger.V(2).Info("protector.Failure(2/2)", "login", login, "id", id)
}

func (p *protector) Exit(id int64, login string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logger.V(2).Info("protector.Exit()", "login", login, "id", id)
}

func (p *protector) clean() {
	p.logger.V(2).Info("protector.clean.tick")
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for k, v := range p.stateByLogin {
		if v.lastFailure.Add(p.cleanDelay).Before(now) {
			p.logger.V(0).Info("protector.clean", "login", k)
			delete(p.stateByLogin, k)
		}
	}
}

func (p *protector) delayFromFailureCount(count int64) time.Duration {
	if count <= p.freeFailure {
		return 0
	}
	penalty := time.Duration(count-p.freeFailure) * p.penaltyByFailure
	if penalty > p.maxPenalty {
		penalty = p.maxPenalty
	}
	return penalty
}

package protector

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"skas/sk-common/proto/v1/proto"
	"sync"
	"sync/atomic"
	"time"
)

// The Protector is a mechanism of protection against BFA.
// It introduce an increasing delay on response in case of failure on a given login
// After some period without failure, the history is cleanup up.
//

type LoginProtector interface {
	EntryForLogin(login string) (locked bool)
	ProtectLoginResult(login string, status proto.Status)
}

type TokenProtector interface {
	EntryForToken() (locked bool)
	TokenNotFound()
}

type Protector interface {
	LoginProtector
	TokenProtector
}

var _ Protector = &protector{}

type loginState struct {
	lastFailure     time.Time
	nbrOfFailure    int64
	pendingFailures atomic.Int64 // Access is NOT protected by some mutex
}

type protector struct {
	mu                sync.Mutex
	logger            logr.Logger
	stateByLogin      map[string]*loginState
	cleanerPeriod     time.Duration
	cleanDelay        time.Duration
	freeFailure       int64 // No delay introduced up to this value
	maxPenalty        time.Duration
	penaltyByFailure  time.Duration
	maxPendingFailure int64
}

const UnknownUser = "_unknownUser_"
const UnknownToken = "_unknownToken_"

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

func WithMaxPendingFailure(mpf int64) Option {
	return func(p *protector) {
		p.maxPendingFailure = mpf
	}
}

// New build a new protector against Brut Force Attack.
// Return nil if !activated. It is up to the caller to test at run time
func New(activated bool, ctx context.Context, logger logr.Logger, opts ...Option) Protector {
	if !activated {
		logger.V(1).Info("Protection NOT activated")
		return &empty{}
	}
	logger.Info("Protection activated")
	p := &protector{
		logger:            logger,
		stateByLogin:      make(map[string]*loginState),
		cleanerPeriod:     60 * time.Second,
		cleanDelay:        30 * time.Minute,
		freeFailure:       4,
		maxPenalty:        15 * time.Second,
		penaltyByFailure:  1 * time.Second,
		maxPendingFailure: 20,
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

func (p *protector) EntryForLogin(login string) bool /*locked*/ {
	p.mu.Lock()
	defer p.mu.Unlock()
	state, ok := p.stateByLogin[login]
	if ok && state.pendingFailures.Load() > p.maxPendingFailure {
		p.logger.V(0).Info("*******WARNING: Too many pending password failing request. May be an attack ", "login", login)
		return true
	}
	state, ok = p.stateByLogin[UnknownUser]
	if ok && state.pendingFailures.Load() > p.maxPendingFailure {
		p.logger.V(0).Info("*******WARNING: Too many pending user failing request. May be an attack ", "login", login)
		return true
	}
	p.logger.V(2).Info("protector.EntryForLogin()", "login", login)
	return false
}

func (p *protector) EntryForToken() bool /*locked*/ {
	p.mu.Lock()
	defer p.mu.Unlock()
	state, ok := p.stateByLogin[UnknownToken]
	if ok && state.pendingFailures.Load() > p.maxPendingFailure {
		p.logger.V(0).Info("*******WARNING: Too many pending token failing request. May be an attack ")
		return true
	}
	p.logger.V(2).Info("protector.EntryForToken()")
	return false
}

func (p *protector) failure(login string) {
	p.logger.V(2).Info("protector.Failure(1/2)", "login", login)
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
	p.logger.V(0).Info("protector.failure", "login", login, "failureCount", nbrOfFailure, "delay", delay.String(), "pendingFailure", state.pendingFailures.Load())
	state.pendingFailures.Add(1)
	time.Sleep(delay)
	state.pendingFailures.Add(-1)
	p.logger.V(2).Info("protector.Failure(2/2)", "login", login)
}

func (p *protector) TokenNotFound() {
	p.failure(UnknownToken)
}

//func (p *protector) ProtectLoginResult(login string, status proto.Status) {
//	if status == proto.UserNotFound {
//		p.failure(UnknownUser)
//	} else if status == proto.PasswordFail || status == proto.InvalidOldPassword {
//		p.failure(login)
//	}
//}

// We removed protection against UserNotFound, as this is a normal case in a multi-provider context

func (p *protector) ProtectLoginResult(login string, status proto.Status) {
	if status == proto.PasswordFail || status == proto.InvalidOldPassword {
		p.failure(login)
	}
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

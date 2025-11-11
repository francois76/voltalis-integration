package scheduler

import (
	"sync"
	"time"
)

// Scheduler exécute périodiquement une fonction et permet un réveil manuel.
// Comportement:
// - Appelle f() immédiatement, puis attend `delay` avant la prochaine exécution.
// - On peut appeler Trigger() pour forcer une exécution immédiate (ignore le délai restant).
// - Trigger annule le sleep courant pour éviter des exécutions rapprochées.
// - Stop arrête proprement le scheduler.
type Scheduler struct {
	delay time.Duration
	f     func() error

	wake chan struct{}
	stop chan struct{}

	wg   sync.WaitGroup
	errc chan error
}

// New crée un nouveau Scheduler. La fonction passée `f` sera exécutée par Start().
func New(delay time.Duration, f func() error) *Scheduler {
	return &Scheduler{
		delay: delay,
		f:     f,
		// buffer 1 pour permettre de déclencher sans bloquer si déjà une demande en attente
		wake:  make(chan struct{}, 1),
		stop:  make(chan struct{}),
		errc:  make(chan error, 1),
	}
}

// Start lance le scheduler dans une goroutine.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.loop()
}

// loop est la boucle principale qui appelle f(), puis attend soit le timer, soit un Trigger, soit Stop.
func (s *Scheduler) loop() {
	defer s.wg.Done()

	for {
		if err := s.f(); err != nil {
			// envoyer l'erreur si possible puis arrêter
			select {
			case s.errc <- err:
			default:
			}
			return
		}

		timer := time.NewTimer(s.delay)

		select {
		case <-timer.C:
			// délai écoulé — on boucle et on appelle f() de nouveau
		case <-s.wake:
			// réveil manuel : annuler le timer pour éviter double-exécution
			if !timer.Stop() {
				// si Stop retourne false, la valeur peut être dans timer.C — la drainer
				select {
				case <-timer.C:
				default:
				}
			}
			// on continue immédiatement
		case <-s.stop:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		}
	}
}

// Trigger réveille le scheduler pour exécuter f() immédiatement. Envoi non-bloquant.
func (s *Scheduler) Trigger() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Stop arrête le scheduler et attend la fin de la goroutine.
func (s *Scheduler) Stop() {
	// fermer stop pour prévenir la boucle
	close(s.stop)
	s.wg.Wait()
}

// Err retourne un canal qui recevra la première erreur rencontrée par f(), si elle survient.
func (s *Scheduler) Err() <-chan error { return s.errc }


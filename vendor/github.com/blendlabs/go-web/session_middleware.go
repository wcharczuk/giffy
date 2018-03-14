package web

// SessionAware is an action that injects the session into the context, it acquires a read lock on session.
func SessionAware(action Action) Action {
	return sessionAware(action, SessionReadLock)
}

// SessionAwareMutating is an action that injects the session into the context and requires a write lock.
func SessionAwareMutating(action Action) Action {
	return sessionAware(action, SessionReadWriteLock)
}

// SessionAwareLockFree is an action that injects the session into the context without acquiring any (read or write) locks.
func SessionAwareLockFree(action Action) Action {
	return sessionAware(action, SessionLockFree)
}

func sessionAware(action Action, sessionLockPolicy int) Action {
	return func(ctx *Ctx) Result {
		session, err := ctx.Auth().VerifySession(ctx)
		if err != nil && !IsErrSessionInvalid(err) {
			return ctx.DefaultResultProvider().InternalError(err)
		}

		if session != nil {
			ctx.SetSession(session)

			switch sessionLockPolicy {
			case SessionReadLock:
				{
					session.RLock()
					defer session.RUnlock()
					break
				}
			case SessionReadWriteLock:
				{
					session.Lock()
					defer session.Unlock()
					break
				}
			}
		}

		return action(ctx)
	}
}

// SessionRequired is an action that requires a session to be present
// or identified in some form on the request, and acquires a read lock on session.
func SessionRequired(action Action) Action {
	return sessionRequired(action, SessionReadLock)
}

// SessionRequiredMutating is an action that requires the session to present and also requires a write lock.
func SessionRequiredMutating(action Action) Action {
	return sessionRequired(action, SessionReadWriteLock)
}

// SessionRequiredLockFree is an action that requires the session to present and does not acquire any (read or write) locks.
func SessionRequiredLockFree(action Action) Action {
	return sessionRequired(action, SessionLockFree)
}

func sessionRequired(action Action, sessionLockPolicy int) Action {
	return func(ctx *Ctx) Result {
		session, err := ctx.Auth().VerifySession(ctx)
		if err != nil && !IsErrSessionInvalid(err) {
			return ctx.DefaultResultProvider().InternalError(err)
		}
		if session == nil {
			return ctx.Auth().Redirect(ctx)
		}

		switch sessionLockPolicy {
		case SessionReadLock:
			{
				session.RLock()
				defer session.RUnlock()
				break
			}
		case SessionReadWriteLock:
			{
				session.Lock()
				defer session.Unlock()
				break
			}
		}

		ctx.SetSession(session)
		return action(ctx)
	}
}

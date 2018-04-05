package web

// SessionAware is an action that injects the session into the context, it acquires a read lock on session.
func SessionAware(action Action) Action {
	return sessionMiddleware(action, nil, SessionReadLock)
}

// SessionAwareMutating is an action that injects the session into the context and requires a write lock.
func SessionAwareMutating(action Action) Action {
	return sessionMiddleware(action, nil, SessionReadWriteLock)
}

// SessionAwareUnsafe is an action that injects the session into the context without acquiring any (read or write) locks.
func SessionAwareUnsafe(action Action) Action {
	return sessionMiddleware(action, nil, SessionUnsafe)
}

// SessionRequired is an action that requires a session to be present
// or identified in some form on the request, and acquires a read lock on session.
func SessionRequired(action Action) Action {
	return sessionMiddleware(action, AuthManagerRedirect, SessionReadLock)
}

// SessionRequiredMutating is an action that requires the session to present and also requires a write lock.
func SessionRequiredMutating(action Action) Action {
	return sessionMiddleware(action, AuthManagerRedirect, SessionReadWriteLock)
}

// SessionRequiredUnsafe is an action that requires the session to present and does not acquire any (read or write) locks.
func SessionRequiredUnsafe(action Action) Action {
	return sessionMiddleware(action, AuthManagerRedirect, SessionUnsafe)
}

// AuthManagerRedirect is a redirect.
func AuthManagerRedirect(ctx *Ctx) Result {
	return ctx.Auth().Redirect(ctx)
}

// SessionMiddleware creates a custom session middleware.
func SessionMiddleware(notAuthorized Action, lockPolicy SessionLockPolicy) Middleware {
	return func(action Action) Action {
		return sessionMiddleware(action, notAuthorized, lockPolicy)
	}
}

// SessionMiddleware returns a session middleware.
func sessionMiddleware(action, notAuthorized Action, lockPolicy SessionLockPolicy) Action {
	return func(ctx *Ctx) Result {
		session, err := ctx.Auth().VerifySession(ctx)
		if err != nil && !IsErrSessionInvalid(err) {
			return ctx.DefaultResultProvider().InternalError(err)
		}

		if session == nil {
			if notAuthorized != nil {
				return notAuthorized(ctx)
			}
			return action(ctx)
		}

		switch lockPolicy {
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

		ctx.WithSession(session)
		return action(ctx)
	}
}

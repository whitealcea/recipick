import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import type { PropsWithChildren } from 'react';
import type { Session } from '@supabase/supabase-js';
import { clearAccessToken, hasStaticAccessToken, setAccessToken } from '../lib/auth';
import { queryClient } from '../lib/queryClient';
import { supabase } from '../lib/supabase';

type AuthContextValue = {
  isLoading: boolean;
  isAuthenticated: boolean;
  userEmail: string | null;
  signInWithGoogle: (nextPath: string) => Promise<void>;
  signOut: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const syncAccessToken = (session: Session | null) => {
  if (session?.access_token) {
    setAccessToken(session.access_token);
    return;
  }
  clearAccessToken();
};

export const AuthProvider = ({ children }: PropsWithChildren) => {
  const [session, setSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (hasStaticAccessToken()) {
      setIsLoading(false);
      return;
    }

    let isMounted = true;

    supabase.auth
      .getSession()
      .then(({ data, error }) => {
        if (!isMounted) return;
        if (error) {
          clearAccessToken();
          setSession(null);
          setIsLoading(false);
          return;
        }
        setSession(data.session ?? null);
        syncAccessToken(data.session ?? null);
        setIsLoading(false);
      })
      .catch(() => {
        if (!isMounted) return;
        clearAccessToken();
        setSession(null);
        setIsLoading(false);
      });

    const { data } = supabase.auth.onAuthStateChange((_event, nextSession) => {
      if (!isMounted) return;
      setSession(nextSession);
      syncAccessToken(nextSession);
      if (!nextSession) {
        queryClient.clear();
      }
    });

    return () => {
      isMounted = false;
      data.subscription.unsubscribe();
    };
  }, []);

  const value = useMemo<AuthContextValue>(() => {
    if (hasStaticAccessToken()) {
      return {
        isLoading: false,
        isAuthenticated: true,
        userEmail: null,
        signInWithGoogle: async () => {
          throw new Error('Static access token mode is enabled');
        },
        signOut: async () => {
          clearAccessToken();
          queryClient.clear();
        },
      };
    }

    return {
      isLoading,
      isAuthenticated: !!session,
      userEmail: session?.user?.email ?? null,
      signInWithGoogle: async (nextPath: string) => {
        const redirectURL = new URL('/login', window.location.origin);
        redirectURL.searchParams.set('next', nextPath || '/recipes');

        const { error } = await supabase.auth.signInWithOAuth({
          provider: 'google',
          options: {
            redirectTo: redirectURL.toString(),
          },
        });
        if (error) {
          throw error;
        }
      },
      signOut: async () => {
        const { error } = await supabase.auth.signOut();
        if (error) {
          throw error;
        }
        clearAccessToken();
        queryClient.clear();
      },
    };
  }, [isLoading, session]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextValue => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used inside AuthProvider');
  }
  return ctx;
};

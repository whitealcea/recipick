import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import { getErrorMessage } from '../lib/errors';

export const LoginPage = () => {
  const [searchParams] = useSearchParams();
  const { signInWithGoogle } = useAuth();

  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const next = searchParams.get('next') || '/recipes';

  return (
    <div className="page login-page">
      <div className="card auth-card">
        <h1>ログイン</h1>
        <p className="muted">
          パスワードを保持しない運用のため、Google OAuth のみを提供しています。
        </p>

        {errorMessage ? <p className="error">{errorMessage}</p> : null}

        <div className="actions">
          <button
            type="button"
            className="oauth-google"
            disabled={isSubmitting}
            onClick={async () => {
              setErrorMessage(null);
              setIsSubmitting(true);
              try {
                await signInWithGoogle(next);
              } catch (error) {
                setErrorMessage(getErrorMessage(error, 'Googleログインに失敗しました。'));
                setIsSubmitting(false);
              }
            }}
          >
            {isSubmitting ? 'Googleへ遷移中...' : 'Googleでログイン'}
          </button>
        </div>
      </div>
    </div>
  );
};

import { useState } from 'react';
import type { ReactElement } from 'react';
import { Link, Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom';
import { useAuth } from './auth/AuthContext';
import { RecipeDetailPage } from './pages/RecipeDetailPage';
import { RecipeEditPage } from './pages/RecipeEditPage';
import { LoginPage } from './pages/LoginPage';
import { RecipeListPage } from './pages/RecipeListPage';
import { RecipeNewPage } from './pages/RecipeNewPage';

const AuthLoadingPage = () => (
  <div className="page">
    <p className="card">認証状態を確認中...</p>
  </div>
);

const RequireAuth = ({ children }: { children: ReactElement }) => {
  const location = useLocation();
  const { isLoading, isAuthenticated } = useAuth();

  if (isLoading) {
    return <AuthLoadingPage />;
  }

  if (!isAuthenticated) {
    const next = `${location.pathname}${location.search}${location.hash}`;
    return <Navigate to={`/login?next=${encodeURIComponent(next)}`} replace />;
  }

  return children;
};

const PublicOnly = ({ children }: { children: ReactElement }) => {
  const location = useLocation();
  const { isLoading, isAuthenticated } = useAuth();

  if (isLoading) {
    return <AuthLoadingPage />;
  }
  if (isAuthenticated) {
    const params = new URLSearchParams(location.search);
    const next = params.get('next') || '/recipes';
    return <Navigate to={next} replace />;
  }
  return children;
};

const ProtectedLayout = () => {
  const { userEmail, signOut } = useAuth();
  const [isSigningOut, setIsSigningOut] = useState(false);

  return (
    <>
      <header className="auth-nav">
        <div className="auth-nav-inner">
          <Link to="/recipes" className="brand-link" aria-label="レシピ一覧へ戻る">
            <img src="/recipick-icon.svg" alt="" className="brand-icon" />
            <span>Recipick</span>
          </Link>

          <div className="auth-nav-actions">
            <span className="muted">{userEmail ? `ログイン中: ${userEmail}` : 'ログイン中'}</span>
            <button
              type="button"
              onClick={async () => {
                setIsSigningOut(true);
                try {
                  await signOut();
                } finally {
                  setIsSigningOut(false);
                }
              }}
              disabled={isSigningOut}
            >
              {isSigningOut ? 'ログアウト中...' : 'ログアウト'}
            </button>
          </div>
        </div>
      </header>
      <Outlet />
    </>
  );
};

const App = () => {
  return (
    <Routes>
      <Route
        path="/login"
        element={
          <PublicOnly>
            <LoginPage />
          </PublicOnly>
        }
      />

      <Route
        element={
          <RequireAuth>
            <ProtectedLayout />
          </RequireAuth>
        }
      >
        <Route path="/" element={<Navigate to="/recipes" replace />} />
        <Route path="/recipes" element={<RecipeListPage />} />
        <Route path="/recipes/new" element={<RecipeNewPage />} />
        <Route path="/recipes/:id" element={<RecipeDetailPage />} />
        <Route path="/recipes/:id/edit" element={<RecipeEditPage />} />
      </Route>
    </Routes>
  );
};

export default App;

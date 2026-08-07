import { Navigate, useLocation } from 'react-router-dom';
import type { ReactNode } from 'react';

const ADMIN_STORAGE_KEY = 'cv_admin';

interface Props {
  children: ReactNode;
}

export default function ProtectedRoute({ children }: Props) {
  const location = useLocation();

  if (localStorage.getItem(ADMIN_STORAGE_KEY) !== 'true') {
    return <Navigate to="/admin/login" replace state={{ from: location.pathname }} />;
  }

  return <>{children}</>;
}

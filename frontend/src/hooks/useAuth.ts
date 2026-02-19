import { useQuery } from '@tanstack/react-query';
import { fetchCurrentUser } from '../api/auth';

export function useAuth() {
  const query = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: fetchCurrentUser,
    retry: false,          // Don't retry 401s
    staleTime: 5 * 60_000, // Cache for 5 minutes
  });

  return {
    user: query.data,
    isLoading: query.isLoading,
    isAuthenticated: query.isSuccess && !!query.data,
    isError: query.isError,
  };
}

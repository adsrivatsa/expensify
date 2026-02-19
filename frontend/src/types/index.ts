export interface User {
  id: string;
  google_id: string;
  email: string;
  name: string;
  picture: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  user_id?: string;
  name: string;
  icon: string;
  color: string;
  is_default: boolean;
  created_at: string;
}

export interface Transaction {
  id: string;
  category_id: string;
  category_name: string;
  category_color: string;
  category_icon: string;
  type: 'inflow' | 'outflow';
  amount: number;
  description: string;
  date: string;
  created_at: string;
  updated_at: string;
}

export interface PaginatedTransactions {
  items: Transaction[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface CreateTransactionPayload {
  category_id: string;
  type: 'inflow' | 'outflow';
  amount: number;
  description: string;
  date: string;
}

export interface UpdateTransactionPayload {
  category_id: string;
  type: 'inflow' | 'outflow';
  amount: number;
  description: string;
  date: string;
}

export interface MonthlyPoint {
  year: number;
  month: number;
  inflow: number;
  outflow: number;
}

export interface CategoryPoint {
  category_id: string;
  category_name: string;
  category_color: string;
  category_icon: string;
  total: number;
}

export interface CashflowSummary {
  monthly: MonthlyPoint[];
  by_category: CategoryPoint[];
}

export interface CreateCategoryPayload {
  name: string;
  icon: string;
  color: string;
}

export interface ApiEnvelope<T> {
  data?: T;
  error?: string;
}

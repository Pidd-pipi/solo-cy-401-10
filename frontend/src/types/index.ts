export * from './enums';

export interface User {
  id: number;
  username: string;
  email?: string;
  name?: string;
  role: string;
  avatar?: string;
  skills?: string[];
  bio?: string;
  contact?: string;
  rating?: number;
  createdAt?: string;
}

export interface Requirement {
  id: number;
  title: string;
  description: string;
  minBudget: number;
  maxBudget: number;
  deadline: string;
  skills: string[];
  status: string;
  publisherId: number;
  winnerId?: number;
  createdAt?: string;
  publisher?: User;
  bids?: Bid[];
}

export interface Bid {
  id: number;
  requirementId: number;
  bidderId: number;
  amount: number;
  durationDays: number;
  proposal: string;
  attachments: string[];
  status: string;
  createdAt?: string;
  bidder?: User;
}

export interface ContractStage {
  name: string;
  amount: number;
  status: string;
  dueAt: string;
}

export interface Contract {
  id: number;
  contractNo: string;
  totalAmount: number;
  paymentType: string;
  stages: ContractStage[];
  status: string;
  requirementId: number;
  partyAId: number;
  partyBId: number;
  createdAt?: string;
  partyA?: User;
  partyB?: User;
  requirement?: Requirement;
}

export interface PageResult<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface DashboardData {
  myRequirements: Requirement[];
  myBids: Bid[];
  myContracts: Contract[];
  counts: Record<string, number>;
}

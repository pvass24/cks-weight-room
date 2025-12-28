// TypeScript types for CKS exercises

export interface Exercise {
  slug: string
  title: string
  description: string
  category: string
  difficulty: 'easy' | 'medium' | 'hard'
  points: number
  estimatedMinutes: number
  prerequisites: string[]
  hints: string[]
  solution: string
}

export interface ExercisesResponse {
  success: boolean
  exercises?: Exercise[]
  exercise?: Exercise
  errorCode?: string
  message?: string
}

export const CategoryLabels: Record<string, string> = {
  'cluster-setup': 'Cluster Setup',
  'cluster-hardening': 'Cluster Hardening',
  'system-hardening': 'System Hardening',
  'minimize-microservice-vulnerabilities': 'Minimize Microservice Vulnerabilities',
  'supply-chain-security': 'Supply Chain Security',
  'monitoring-logging-runtime-security': 'Monitoring, Logging & Runtime Security',
}

export const DifficultyColors: Record<string, string> = {
  easy: 'text-green-600 bg-green-50',
  medium: 'text-yellow-600 bg-yellow-50',
  hard: 'text-red-600 bg-red-50',
}

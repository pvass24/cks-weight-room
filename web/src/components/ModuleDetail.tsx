'use client'

import { useRouter } from 'next/navigation'
import AppLayout from '@/components/AppLayout'
import { ArrowLeft, BookOpen, Clock, CheckCircle2 } from 'lucide-react'

const moduleData: Record<string, any> = {
  'cluster-setup': {
    title: 'Cluster Setup',
    description: 'Learn to set up secure Kubernetes clusters with proper network policies',
    icon: '📘',
    sessions: [
      { id: 1, title: 'Kubernetes Architecture Overview', duration: 30, completed: true },
      { id: 2, title: 'Cluster Installation & Configuration', duration: 45, completed: true },
      { id: 3, title: 'Network Policies & CNI', duration: 40, completed: false },
      { id: 4, title: 'Cluster Security Best Practices', duration: 35, completed: false },
    ]
  },
  'cluster-hardening': {
    title: 'Cluster Hardening',
    description: 'Harden your Kubernetes cluster with RBAC, service accounts, and security contexts',
    icon: '📗',
    sessions: [
      { id: 1, title: 'RBAC & Service Accounts', duration: 40, completed: false },
      { id: 2, title: 'Pod Security Standards', duration: 35, completed: false },
      { id: 3, title: 'Network Security', duration: 45, completed: false },
      { id: 4, title: 'API Server Security', duration: 30, completed: false },
    ]
  },
  'system-hardening': {
    title: 'System Hardening',
    description: 'Secure the underlying system with AppArmor, Seccomp, and proper host security',
    icon: '📙',
    sessions: [
      { id: 1, title: 'AppArmor Profiles', duration: 40, completed: false },
      { id: 2, title: 'Seccomp Profiles', duration: 35, completed: false },
      { id: 3, title: 'Host Security', duration: 30, completed: false },
      { id: 4, title: 'Kernel Hardening', duration: 35, completed: false },
    ]
  },
  'microservice-security': {
    title: 'Microservice Security',
    description: 'Secure microservices with proper isolation and security policies',
    icon: '📕',
    sessions: [
      { id: 1, title: 'Service Mesh Security', duration: 45, completed: false },
      { id: 2, title: 'mTLS Configuration', duration: 40, completed: false },
      { id: 3, title: 'API Gateway Security', duration: 35, completed: false },
      { id: 4, title: 'Zero Trust Architecture', duration: 40, completed: false },
    ]
  },
  'supply-chain': {
    title: 'Supply Chain Security',
    description: 'Secure your container images and implement image scanning and signing',
    icon: '📔',
    sessions: [
      { id: 1, title: 'Image Scanning with Trivy', duration: 35, completed: false },
      { id: 2, title: 'Image Signing & Verification', duration: 40, completed: false },
      { id: 3, title: 'Admission Controllers', duration: 45, completed: false },
      { id: 4, title: 'Software Bill of Materials', duration: 30, completed: false },
    ]
  },
  'monitoring': {
    title: 'Monitoring & Runtime',
    description: 'Monitor and secure runtime behavior with Falco and audit logging',
    icon: '📓',
    sessions: [
      { id: 1, title: 'Falco Runtime Security', duration: 45, completed: false },
      { id: 2, title: 'Audit Logging', duration: 35, completed: false },
      { id: 3, title: 'Threat Detection', duration: 40, completed: false },
      { id: 4, title: 'Incident Response', duration: 35, completed: false },
    ]
  }
}

interface ModuleDetailProps {
  moduleId: string
}

export default function ModuleDetail({ moduleId }: ModuleDetailProps) {
  const router = useRouter()
  const module = moduleData[moduleId]

  if (!module) {
    return (
      <AppLayout>
        <div className="p-6">
          <div className="text-center py-12">
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">Module Not Found</h1>
            <p className="text-slate-600 dark:text-slate-400 mb-4">The requested module does not exist.</p>
            <button
              onClick={() => router.push('/modules')}
              className="text-blue-600 dark:text-blue-400 hover:underline"
            >
              Back to Modules
            </button>
          </div>
        </div>
      </AppLayout>
    )
  }

  const completedSessions = module.sessions.filter((s: any) => s.completed).length
  const progress = Math.round((completedSessions / module.sessions.length) * 100)

  return (
    <AppLayout>
      <div className="p-6 space-y-6">
        {/* Header */}
        <div>
          <button
            onClick={() => router.push('/modules')}
            className="flex items-center gap-2 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white mb-4"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Back to Modules</span>
          </button>

          <div className="flex items-start gap-4">
            <div className="text-5xl">{module.icon}</div>
            <div className="flex-1">
              <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">
                {module.title}
              </h1>
              <p className="text-slate-600 dark:text-slate-400 mb-4">
                {module.description}
              </p>
              <div className="flex items-center gap-6">
                <div className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400">
                  <BookOpen className="w-4 h-4" />
                  <span>{module.sessions.length} sessions</span>
                </div>
                <div className="text-sm font-semibold text-slate-900 dark:text-white">
                  {progress}% Complete
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Progress Bar */}
        <div className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-slate-900 dark:text-white">Overall Progress</span>
            <span className="text-sm font-semibold text-slate-900 dark:text-white">{progress}%</span>
          </div>
          <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2.5">
            <div
              className="bg-blue-600 h-2.5 rounded-full transition-all duration-500"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>

        {/* Sessions List */}
        <div className="space-y-3">
          {module.sessions.map((session: any, idx: number) => (
            <div
              key={session.id}
              className="bg-white dark:bg-slate-800 rounded-xl p-5 border border-slate-200 dark:border-slate-700 hover:shadow-md transition-shadow"
            >
              <div className="flex items-center gap-4">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
                  session.completed
                    ? 'bg-green-100 dark:bg-green-900/20'
                    : 'bg-slate-100 dark:bg-slate-700'
                }`}>
                  {session.completed ? (
                    <CheckCircle2 className="w-5 h-5 text-green-600 dark:text-green-400" />
                  ) : (
                    <span className="text-sm font-semibold text-slate-600 dark:text-slate-400">
                      {idx + 1}
                    </span>
                  )}
                </div>

                <div className="flex-1">
                  <h3 className="font-semibold text-slate-900 dark:text-white mb-1">
                    {session.title}
                  </h3>
                  <div className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400">
                    <Clock className="w-4 h-4" />
                    <span>{session.duration} minutes</span>
                  </div>
                </div>

                <button className="px-4 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition-colors">
                  {session.completed ? 'Review' : 'Start'}
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </AppLayout>
  )
}

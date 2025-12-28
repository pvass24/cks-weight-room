'use client'

import AppLayout from '@/components/AppLayout'
import { useRouter } from 'next/navigation'
import { BookOpen, ArrowRight, CheckCircle } from 'lucide-react'

interface Module {
  id: string
  title: string
  description: string
  sessions: number
  progress: number
  color: string
  icon: string
}

export default function ModulesPage() {
  const router = useRouter()

  const modules: Module[] = [
    {
      id: 'cluster-setup',
      title: 'Cluster Setup',
      description: 'Learn to set up secure Kubernetes clusters with proper network policies',
      sessions: 4,
      progress: 60,
      color: 'bg-blue-500',
      icon: '📘'
    },
    {
      id: 'cluster-hardening',
      title: 'Cluster Hardening',
      description: 'Harden your Kubernetes cluster with RBAC, service accounts, and security contexts',
      sessions: 4,
      progress: 45,
      color: 'bg-green-500',
      icon: '📗'
    },
    {
      id: 'system-hardening',
      title: 'System Hardening',
      description: 'Secure the underlying system with AppArmor, Seccomp, and proper host security',
      sessions: 4,
      progress: 30,
      color: 'bg-yellow-500',
      icon: '📙'
    },
    {
      id: 'microservice-security',
      title: 'Microservice Security',
      description: 'Secure microservices with proper isolation and security policies',
      sessions: 4,
      progress: 40,
      color: 'bg-orange-500',
      icon: '📕'
    },
    {
      id: 'supply-chain',
      title: 'Supply Chain Security',
      description: 'Secure your container images and implement image scanning and signing',
      sessions: 4,
      progress: 55,
      color: 'bg-red-500',
      icon: '📔'
    },
    {
      id: 'monitoring',
      title: 'Monitoring & Runtime',
      description: 'Monitor and secure runtime behavior with Falco and audit logging',
      sessions: 4,
      progress: 35,
      color: 'bg-purple-500',
      icon: '📓'
    }
  ]

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Study Modules</h1>
          <p className="text-gray-600">
            Comprehensive curriculum covering all CKS exam domains
          </p>
        </div>

        {/* Modules Grid */}
        <div className="space-y-4">
          {modules.map((module) => (
            <div
              key={module.id}
              className="bg-white rounded-xl border border-gray-200 p-6 hover:shadow-lg transition-shadow cursor-pointer group"
              onClick={() => router.push(`/modules/${module.id}`)}
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-4 flex-1">
                  {/* Icon */}
                  <div className="text-4xl">{module.icon}</div>

                  {/* Content */}
                  <div className="flex-1">
                    <h3 className="text-xl font-bold text-gray-900 mb-2 group-hover:text-blue-600 transition-colors">
                      {module.title}
                    </h3>
                    <p className="text-gray-600 mb-4">{module.description}</p>

                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2 text-sm text-gray-600">
                        <BookOpen className="w-4 h-4" />
                        <span>{module.sessions} sessions</span>
                      </div>

                      {/* Progress Bar */}
                      <div className="flex-1 max-w-xs">
                        <div className="flex items-center justify-between mb-1">
                          <span className="text-xs font-medium text-gray-600">Progress</span>
                          <span className="text-xs font-medium text-gray-900">{module.progress}%</span>
                        </div>
                        <div className="w-full bg-gray-200 rounded-full h-2">
                          <div
                            className={`${module.color} h-2 rounded-full transition-all duration-500`}
                            style={{ width: `${module.progress}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Arrow */}
                <ArrowRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600 group-hover:translate-x-1 transition-all" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </AppLayout>
  )
}

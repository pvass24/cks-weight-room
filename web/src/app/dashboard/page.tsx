'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import AppLayout from '@/components/AppLayout'
import {
  TrendingUp,
  Clock,
  Award,
  Target,
  BookOpen,
  CheckCircle2,
  ArrowRight,
} from 'lucide-react'

export default function DashboardPage() {
  const router = useRouter()
  const [stats] = useState({
    totalLessons: 48,
    completedLessons: 0,
    studyHours: 0,
    examAttempts: 0,
  })

  const progressPercentage = Math.round((stats.completedLessons / stats.totalLessons) * 100)

  const domains = [
    { name: 'Cluster Setup', progress: 60, lessons: 8, color: 'bg-blue-600' },
    { name: 'Cluster Hardening', progress: 45, lessons: 10, color: 'bg-green-600' },
    { name: 'System Hardening', progress: 30, lessons: 8, color: 'bg-yellow-600' },
    { name: 'Microservice Security', progress: 40, lessons: 9, color: 'bg-orange-600' },
    { name: 'Supply Chain Security', progress: 55, lessons: 7, color: 'bg-red-600' },
    { name: 'Monitoring & Runtime', progress: 35, lessons: 6, color: 'bg-teal-600' },
  ]

  return (
    <AppLayout>
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
            Welcome Back!
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Continue your journey to becoming a Kubernetes Security Specialist
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-white dark:bg-slate-800 rounded-xl p-5 border border-slate-200 dark:border-slate-700">
            <div className="flex items-start justify-between mb-3">
              <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                <BookOpen className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <span className="text-xs font-medium text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-2 py-1 rounded">
                +12%
              </span>
            </div>
            <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-1">
              {stats.completedLessons}/{stats.totalLessons}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">Lessons Completed</p>
          </div>

          <div className="bg-white dark:bg-slate-800 rounded-xl p-5 border border-slate-200 dark:border-slate-700">
            <div className="flex items-start justify-between mb-3">
              <div className="p-2 bg-green-50 dark:bg-green-900/20 rounded-lg">
                <Clock className="w-5 h-5 text-green-600 dark:text-green-400" />
              </div>
              <span className="text-xs font-medium text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-2 py-1 rounded">
                +5h
              </span>
            </div>
            <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-1">
              {stats.studyHours}h
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">Study Hours</p>
          </div>

          <div className="bg-white dark:bg-slate-800 rounded-xl p-5 border border-slate-200 dark:border-slate-700">
            <div className="flex items-start justify-between mb-3">
              <div className="p-2 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                <Target className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
              </div>
            </div>
            <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-1">
              {progressPercentage}%
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">Overall Progress</p>
          </div>

          <div className="bg-white dark:bg-slate-800 rounded-xl p-5 border border-slate-200 dark:border-slate-700">
            <div className="flex items-start justify-between mb-3">
              <div className="p-2 bg-red-50 dark:bg-red-900/20 rounded-lg">
                <Award className="w-5 h-5 text-red-600 dark:text-red-400" />
              </div>
            </div>
            <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-1">
              {stats.examAttempts}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">Mock Exams Taken</p>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-slate-900 dark:text-white">
                Learning Progress by Domain
              </h2>
              <button className="text-sm text-blue-600 dark:text-blue-400 hover:underline">
                View All
              </button>
            </div>

            <div className="space-y-5">
              {domains.map((domain) => (
                <div key={domain.name}>
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex-1">
                      <h3 className="text-sm font-medium text-slate-900 dark:text-white mb-1">
                        {domain.name}
                      </h3>
                      <p className="text-xs text-slate-500 dark:text-slate-400">
                        {domain.lessons} lessons
                      </p>
                    </div>
                    <span className="text-sm font-semibold text-slate-900 dark:text-white">
                      {domain.progress}%
                    </span>
                  </div>
                  <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2.5">
                    <div
                      className={`${domain.color} h-2.5 rounded-full transition-all duration-500`}
                      style={{ width: `${domain.progress}%` }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="space-y-6">
            <div className="bg-gradient-to-br from-blue-600 to-blue-700 rounded-xl p-6 text-white">
              <TrendingUp className="w-10 h-10 mb-4 opacity-90" />
              <h3 className="text-xl font-bold mb-2">Keep Going!</h3>
              <p className="text-blue-100 text-sm mb-4">
                You're making great progress. Complete 3 more lessons to reach your weekly goal.
              </p>
              <button
                onClick={() => router.push('/modules')}
                className="w-full bg-white text-blue-600 font-medium py-2 px-4 rounded-lg hover:bg-blue-50 transition-colors flex items-center justify-center gap-2"
              >
                Continue Learning
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>

            <div className="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700">
              <h3 className="font-semibold text-slate-900 dark:text-white mb-4">
                Recent Activity
              </h3>
              <div className="space-y-3">
                {[
                  { title: 'Pod Security Standards', time: '2 hours ago', completed: true },
                  { title: 'Network Policies Deep Dive', time: '1 day ago', completed: true },
                  { title: 'Mock Exam Attempt #2', time: '2 days ago', completed: false },
                ].map((activity, idx) => (
                  <div key={idx} className="flex items-start gap-3">
                    <div
                      className={`p-1 rounded-full ${
                        activity.completed
                          ? 'bg-green-100 dark:bg-green-900/20'
                          : 'bg-slate-100 dark:bg-slate-700'
                      }`}
                    >
                      <CheckCircle2
                        className={`w-4 h-4 ${
                          activity.completed
                            ? 'text-green-600 dark:text-green-400'
                            : 'text-slate-400'
                        }`}
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-slate-900 dark:text-white">
                        {activity.title}
                      </p>
                      <p className="text-xs text-slate-500 dark:text-slate-400">{activity.time}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </AppLayout>
  )
}

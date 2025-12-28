'use client'

import { useEffect, useState } from 'react'

interface ProgressStats {
  scenariosCompleted: number
  totalScenarios: number
  completionPercentage: number
  totalPracticeMinutes: number
  averageScore: number
  mockExamsTaken: number
  mockExamsPassed: number
  progressByDomain: DomainProgress[]
  recentActivity: RecentAttempt[]
}

interface DomainProgress {
  domain: string
  displayName: string
  weight: number
  completedCount: number
  totalCount: number
  completionPercentage: number
}

interface RecentAttempt {
  exerciseSlug: string
  exerciseTitle: string
  completedAt: string
  durationSeconds: number
  score: number
  maxScore: number
  passed: boolean
  isPersonalBest: boolean
}

export default function ProgressStats() {
  const [stats, setStats] = useState<ProgressStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [showDomains, setShowDomains] = useState(false)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch('/api/progress/stats')
        if (response.ok) {
          const data = await response.json()
          setStats(data)
        }
      } catch (error) {
        console.error('Failed to fetch progress stats:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchStats()
  }, [])

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="h-20 bg-gray-200 rounded"></div>
        </div>
      </div>
    )
  }

  if (!stats) {
    return null
  }

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${minutes}:${secs.toString().padStart(2, '0')}`
  }

  const formatTime = (minutes: number): string => {
    const hours = Math.floor(minutes / 60)
    const mins = minutes % 60
    if (hours > 0) {
      return `${hours}h ${mins}m`
    }
    return `${mins}m`
  }

  const formatDate = (dateStr: string): string => {
    try {
      const date = new Date(dateStr)
      const now = new Date()
      const diffMs = now.getTime() - date.getTime()
      const diffMins = Math.floor(diffMs / 60000)
      const diffHours = Math.floor(diffMins / 60)
      const diffDays = Math.floor(diffHours / 24)

      if (diffMins < 60) {
        return `${diffMins}m ago`
      } else if (diffHours < 24) {
        return `${diffHours}h ago`
      } else if (diffDays === 1) {
        return 'Yesterday'
      } else if (diffDays < 7) {
        return `${diffDays} days ago`
      } else {
        return date.toLocaleDateString()
      }
    } catch {
      return dateStr
    }
  }

  return (
    <div className="bg-white rounded-lg shadow-md p-6 mb-6">
      <h2 className="text-2xl font-bold text-gray-900 mb-4">Your Progress</h2>

      {/* Overall Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-blue-50 rounded-lg p-4">
          <div className="text-sm text-blue-600 font-medium">Scenarios Completed</div>
          <div className="text-2xl font-bold text-blue-900">
            {stats.scenariosCompleted}/{stats.totalScenarios}
          </div>
          <div className="text-xs text-blue-600">
            {stats.completionPercentage.toFixed(0)}%
          </div>
        </div>

        <div className="bg-green-50 rounded-lg p-4">
          <div className="text-sm text-green-600 font-medium">Practice Time</div>
          <div className="text-2xl font-bold text-green-900">
            {formatTime(stats.totalPracticeMinutes)}
          </div>
          <div className="text-xs text-green-600">&nbsp;</div>
        </div>

        <div className="bg-purple-50 rounded-lg p-4">
          <div className="text-sm text-purple-600 font-medium">Avg Score</div>
          <div className="text-2xl font-bold text-purple-900">
            {stats.averageScore > 0 ? stats.averageScore.toFixed(0) : '—'}%
          </div>
          <div className="text-xs text-purple-600">&nbsp;</div>
        </div>

        <div className="bg-orange-50 rounded-lg p-4">
          <div className="text-sm text-orange-600 font-medium">Mock Exams</div>
          <div className="text-2xl font-bold text-orange-900">
            {stats.mockExamsTaken}
          </div>
          <div className="text-xs text-orange-600">
            {stats.mockExamsPassed} passed
          </div>
        </div>
      </div>

      {/* Progress by Domain - Collapsible */}
      {stats.progressByDomain.length > 0 && (
        <div className="mb-6">
          <button
            onClick={() => setShowDomains(!showDomains)}
            className="w-full flex items-center justify-between text-left font-semibold text-gray-900 mb-3"
          >
            <span>Progress by Domain</span>
            <svg
              className={`w-5 h-5 flex-shrink-0 transition-transform ${showDomains ? 'rotate-180' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          {showDomains && (
            <div className="space-y-3">
              {stats.progressByDomain.map((domain) => (
                <div key={domain.domain} className="border border-gray-200 rounded-lg p-3">
                  <div className="flex items-center justify-between mb-2">
                    <div>
                      <div className="font-medium text-gray-900">{domain.displayName}</div>
                      <div className="text-xs text-gray-500">{domain.weight}% of exam</div>
                    </div>
                    <div className="text-right">
                      <div className="text-lg font-bold text-gray-900">
                        {domain.completedCount}/{domain.totalCount}
                      </div>
                      <div className="text-xs text-gray-600">
                        {domain.completionPercentage.toFixed(0)}%
                      </div>
                    </div>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all"
                      style={{ width: `${domain.completionPercentage}%` }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Recent Activity */}
      {stats.recentActivity.length > 0 && (
        <div>
          <h3 className="font-semibold text-gray-900 mb-3">Recent Activity</h3>
          <div className="space-y-2">
            {stats.recentActivity.map((attempt, idx) => (
              <div key={idx} className="flex items-center justify-between border-l-4 border-gray-300 pl-3 py-2">
                <div className="flex-1">
                  <div className="font-medium text-gray-900">{attempt.exerciseTitle}</div>
                  <div className="text-xs text-gray-500">{formatDate(attempt.completedAt)}</div>
                </div>
                <div className="text-right">
                  <div className={`text-sm font-semibold ${attempt.passed ? 'text-green-600' : 'text-red-600'}`}>
                    {attempt.passed ? 'PASS' : 'FAIL'} • {formatDuration(attempt.durationSeconds)}
                  </div>
                  <div className="text-xs text-gray-600">
                    {attempt.score}/{attempt.maxScore} pts
                    {attempt.isPersonalBest && <span className="ml-2 text-yellow-600">⭐ PB!</span>}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

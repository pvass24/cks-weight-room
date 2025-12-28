'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  BookOpen,
  FlaskConical,
  FileCheck,
  BarChart3,
  Bookmark,
  Shield
} from 'lucide-react'

const navItems = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard, href: '/dashboard' },
  { id: 'modules', label: 'Study Modules', icon: BookOpen, href: '/modules' },
  { id: 'labs', label: 'Practice Labs', icon: FlaskConical, href: '/exercises' },
  { id: 'exams', label: 'Mock Exams', icon: FileCheck, href: '/exams' },
  { id: 'bookmarks', label: 'Bookmarks', icon: Bookmark, href: '/bookmarks' },
  { id: 'analytics', label: 'Analytics', icon: BarChart3, href: '/analytics' },
]

export default function Sidebar() {
  const pathname = usePathname()

  return (
    <aside className="fixed lg:static inset-y-0 left-0 z-50 w-64 bg-white dark:bg-slate-800 border-r border-slate-200 dark:border-slate-700 transform transition-transform duration-200 translate-x-0">
      <div className="flex flex-col h-full">
        <div className="p-6 border-b border-slate-200 dark:border-slate-700">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center">
              <Shield className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="font-bold text-slate-900 dark:text-white">CKS Platform</h1>
              <p className="text-xs text-slate-500 dark:text-slate-400">Security Specialist</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = pathname === item.href || pathname?.startsWith(item.href + '/')
            return (
              <Link
                key={item.id}
                href={item.href}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg font-medium transition-all ${
                  isActive
                    ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400'
                    : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700'
                }`}
              >
                <Icon className="w-5 h-5 flex-shrink-0" />
                <span>{item.label}</span>
              </Link>
            )
          })}
        </nav>

        <div className="p-4 border-t border-slate-200 dark:border-slate-700">
          <div className="bg-gradient-to-br from-blue-50 to-slate-50 dark:from-blue-900/20 dark:to-slate-800 p-4 rounded-lg">
            <h3 className="font-semibold text-sm text-slate-900 dark:text-white mb-1">
              Exam Preparation
            </h3>
            <p className="text-xs text-slate-600 dark:text-slate-400 mb-3">
              Complete all modules to be exam ready
            </p>
            <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
              <div className="bg-blue-600 h-2 rounded-full" style={{ width: '45%' }}></div>
            </div>
            <p className="text-xs text-slate-600 dark:text-slate-400 mt-2">45% Complete</p>
          </div>
        </div>
      </div>
    </aside>
  )
}

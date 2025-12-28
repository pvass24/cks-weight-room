'use client'

import AppLayout from '@/components/AppLayout'
import { Bookmark } from 'lucide-react'

export default function BookmarksPage() {
  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Bookmarks</h1>
          <p className="text-gray-600">
            Save and revisit your favorite lessons and exercises
          </p>
        </div>

        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <Bookmark className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">No bookmarks yet</p>
            <p className="text-sm text-gray-400 mt-2">
              Bookmark lessons and labs to access them quickly
            </p>
          </div>
        </div>
      </div>
    </AppLayout>
  )
}

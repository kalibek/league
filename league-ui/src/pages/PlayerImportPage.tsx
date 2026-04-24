import { Link } from 'react-router-dom'
import { importCSV } from '../api/players'
import { CSVImport } from '../components/CSVImport/CSVImport'
import { extractErrorMessage } from '../hooks/utils'

export function PlayerImportPage() {
  const handleImport = async (file: File) => {
    try {
      const res = await importCSV(file)
      return res.data
    } catch (e) {
      return { imported: 0, skipped: 0, errors: [extractErrorMessage(e)] }
    }
  }

  return (
    <div className="max-w-2xl mx-auto py-8 px-4">
      <Link to="/players" className="text-sm text-blue-600 hover:underline mb-4 block">
        &larr; Back to Players
      </Link>
      <h1 className="text-2xl font-bold text-gray-900 mb-2">Import Players from CSV</h1>
      <p className="text-sm text-gray-500 mb-6">
        Upload a CSV file with columns: <code className="bg-gray-100 px-1 rounded">first_name</code>,{' '}
        <code className="bg-gray-100 px-1 rounded">last_name</code>,{' '}
        <code className="bg-gray-100 px-1 rounded">email</code>,{' '}
        <code className="bg-gray-100 px-1 rounded">initial_rating</code> (optional).
      </p>
      <CSVImport onImport={handleImport} />
    </div>
  )
}

export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8 max-w-md w-full text-center">
        <h2 className="text-2xl font-bold text-red-600 mb-4">404 - Page Not Found</h2>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          The page you are looking for does not exist.
        </p>
        <a
          href="/"
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors inline-block"
        >
          Go Home
        </a>
      </div>
    </div>
  );
}
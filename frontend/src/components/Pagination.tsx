interface Props {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export default function Pagination({ currentPage, totalPages, onPageChange }: Props) {
  if (totalPages <= 1) return null;

  return (
    <div className="flex items-center justify-center gap-2">
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage <= 1}
        className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium
                   text-gray-600 hover:bg-gray-50 disabled:opacity-30 disabled:cursor-not-allowed"
      >
        Previous
      </button>

      <span className="px-4 text-sm text-gray-500">
        Page {currentPage} of {totalPages}
      </span>

      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage >= totalPages}
        className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium
                   text-gray-600 hover:bg-gray-50 disabled:opacity-30 disabled:cursor-not-allowed"
      >
        Next
      </button>
    </div>
  );
}

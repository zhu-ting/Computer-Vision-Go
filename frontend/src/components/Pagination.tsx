interface Props {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export default function Pagination({ currentPage, totalPages, onPageChange }: Props) {
  if (totalPages <= 1) return null;

  return (
    <div className="flex items-center justify-center gap-1 sm:gap-2">
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage <= 1}
        className="min-h-[44px] min-w-[44px] rounded-lg border border-gray-200 px-2 sm:px-3
                   py-2 text-xs sm:text-sm font-medium text-gray-600 hover:bg-gray-50
                   active:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed
                   transition-colors"
      >
        Previous
      </button>

      <span className="px-2 sm:px-4 text-xs sm:text-sm text-gray-500 whitespace-nowrap">
        {currentPage} / {totalPages}
      </span>

      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage >= totalPages}
        className="min-h-[44px] min-w-[44px] rounded-lg border border-gray-200 px-2 sm:px-3
                   py-2 text-xs sm:text-sm font-medium text-gray-600 hover:bg-gray-50
                   active:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed
                   transition-colors"
      >
        Next
      </button>
    </div>
  );
}

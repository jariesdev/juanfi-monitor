	export interface Filter {
		[key: string]: string | number | boolean | null | undefined;
	}

	export interface QueryParameters extends Filter {
		q?: string;
		page: number;
		size: number;
	}

	export interface TableHeader {
		label: string;
		field: string;
	}

	export interface SimpleTableHeader extends TableHeader {
		align?: 'left' | 'center' | 'right';
		sortable?: boolean;
	}

	export interface RowItem {
		[key: string]: any;
	}
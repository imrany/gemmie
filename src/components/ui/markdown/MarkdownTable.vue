<script setup lang="ts">
interface Props {
    data: {
        id: string;
        headers: string[];
        alignments: ("left" | "right" | "center")[];
        rows: string[][];
    };
}

defineProps<Props>();
</script>

<template>
    <div class="table-container my-4 overflow-x-auto custom-scrollbar">
        <table
            class="w-full border-collapse border border-gray-300 dark:border-gray-600"
        >
            <thead class="bg-gray-100 dark:bg-gray-800">
                <tr>
                    <th
                        v-for="(header, index) in data.headers"
                        :key="index"
                        :class="`border border-gray-300 dark:border-gray-600 px-4 py-2 text-${data.alignments[index]} font-semibold text-gray-900 dark:text-gray-100`"
                    >
                        {{ header }}
                    </th>
                </tr>
            </thead>
            <tbody>
                <tr
                    v-for="(row, rowIndex) in data.rows"
                    :key="rowIndex"
                    class="hover:bg-gray-50 dark:hover:bg-gray-700/50"
                >
                    <td
                        v-for="(cell, cellIndex) in row"
                        :key="cellIndex"
                        :class="`border border-gray-300 dark:border-gray-600 px-4 py-2 text-${data.alignments[cellIndex]} text-gray-800 dark:text-gray-200`"
                    >
                        {{ cell }}
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

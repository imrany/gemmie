<script lang="ts" setup>
const plans = [
  {
    name: "Student",
    price: "50 Ksh",
    duration: "per 5 hours",
    description: "Perfect for quick sessions and light academic use.",
    features: ["Strong privacy", "Data sync enabled", "Good for light work"],
    popular: false,
  },
  {
    name: "Pro",
    price: "100 Ksh",
    duration: "per 24 hours",
    description: "Great for professionals needing reliable AI help all day.",
    features: ["Full privacy", "Sync across devices", "Handles heavy workloads"],
    popular: true,
  },
  {
    name: "Hobbyist",
    price: "500 Ksh",
    duration: "per week",
    description: "Best for hobbyists and regular users exploring AI deeply.",
    features: [
      "Top privacy",
      "Persistent sync",
      "Extended access for projects",
    ],
    popular: false,
  },
];
</script>

<template>
  <div class="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-7xl mx-auto text-center mb-12">
      <h1 class="text-3xl font-bold text-gray-900 sm:text-4xl">
        Choose Your Plan
      </h1>
      <p class="mt-4 text-gray-600">
        Flexible pricing designed for Students, Professionals, and Hobbyists.
      </p>
    </div>

    <div class="grid gap-8 md:grid-cols-3 max-w-7xl mx-auto">
      <div
        v-for="plan in plans"
        :key="plan.name"
        class="flex flex-col bg-white border rounded-2xl shadow-sm hover:shadow-lg transition-all duration-200 overflow-hidden"
        :class="plan.popular ? 'border-blue-600 ring-2 ring-blue-600' : ''"
      >
        <div class="p-6 flex-grow flex flex-col">
          <h2 class="text-xl font-semibold text-gray-900 mb-2">
            {{ plan.name }}
          </h2>
          <p class="text-gray-600 mb-4">{{ plan.description }}</p>

          <div class="mb-6">
            <span class="text-3xl font-bold text-gray-900">{{ plan.price }}</span>
            <span class="text-gray-500 text-sm ml-1">{{ plan.duration }}</span>
          </div>

          <ul class="space-y-2 flex-grow">
            <li
              v-for="feature in plan.features"
              :key="feature"
              class="flex items-center text-gray-700 text-sm"
            >
              <i class="pi pi-check text-green-600 mr-2"></i>
              <span>{{ feature }}</span>
            </li>
          </ul>

          <button
            class="mt-6 w-full py-2 px-4 rounded-lg text-white font-medium transition-all"
            :class="
              plan.popular
                ? 'bg-blue-600 hover:bg-blue-700'
                : 'bg-gray-800 hover:bg-gray-900'
            "
          >
            Get Started
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

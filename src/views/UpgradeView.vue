<script lang="ts" setup>
import { onMounted, ref, watch, computed, inject, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import { useRouter } from 'vue-router';
import { toast } from 'vue-sonner';

const route = useRoute()
const router = useRouter()
const planName = route.params.plan as 'student' | 'pro' | 'hobbyist' | undefined
const selectPlanName = ref<'student' | 'pro' | 'hobbyist'>(planName ?? 'pro')
const {
  parsedUserDetails
}=inject("globalState") as {
  parsedUserDetails: any
}

const plans = ref([
  {
    name: "Student",
    id: "student",
    price: 50,
    duration: "per 5 hours",
    description: "Perfect for quick sessions and light academic use.",
    features: [
      "Strong privacy", 
      "Good for light work",
      "Analyze and summarize text",
      "Write, edit and create content", 
      "Get web results and insights", 
    ],
    popular: false,
  },
  {
    name: "Pro",
    id: "pro",
    price: 100,
    duration: "per 24 hours",
    description: "Great for professionals needing reliable AI help all day.",
    features: [
      "Everything in Student Plan", 
      "Handles heavy workloads",
      "Downloadable content e.g PDFs", 
      "Full privacy", 
      "Data sync across all devices", 
    ],
    popular: true,
  },
  {
    name: "Hobbyist",
    id: "hobbyist",
    price: 500,
    duration: "per week",
    description: "Best for hobbyists and regular users exploring AI deeply.",
    features: [
      "Everything in Pro Plan", 
      "More usage time",
      "Persistent sync",
      "Extended access for projects",
    ],
    popular: false,
  },
])

const showCheckout = ref(false)
const isProcessing = ref(false)

// Payment form data - create local reactive references
const paymentForm = ref({
  username: '',
  email: '',
  phone: '',
})

// Form validation
const isFormValid = computed(() => {
  const phoneRegex = /^(\+254|0)[17][0-9]{8}$/
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  
  // Check username
  const hasValidUsername = paymentForm.value.username.trim() !== ''
  
  // Check email
  const hasValidEmail = emailRegex.test(paymentForm.value.email)
  
  // Check phone
  const hasValidPhone = phoneRegex.test(paymentForm.value.phone)
  
  return hasValidUsername && hasValidEmail && hasValidPhone
})

const expiryTimestamp = ref<number | null>(null)
const now = ref(Date.now())
let timer: number | null = null

// Compute expiry date string - updates in real-time
const expiryDate = computed(() => {
  if (!expiryTimestamp.value) return ''
  const futureTime = now.value + expiryTimestamp.value
  return new Date(futureTime).toLocaleString('en-KE', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
})

// Reactive remaining time
const timeLeft = computed(() => {
  if (!expiryTimestamp.value) return ''
  const diff = expiryTimestamp.value
  if (diff <= 0) return 'Expired'
  
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((diff % (1000 * 60)) / 1000)
  
  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`
  }
  return `${hours}h ${minutes}m ${seconds}s`
})

// Set expiry timestamp based on selected plan - updates in real-time
function setExpiry(planId: string) {
  // This will be recalculated reactively based on current time
  switch (planId) {
    case 'student':
      expiryTimestamp.value = 5 * 60 * 60 * 1000 // 5 hours in milliseconds
      break
    case 'pro':
      expiryTimestamp.value = 24 * 60 * 60 * 1000 // 24 hours in milliseconds
      break
    case 'hobbyist':
      expiryTimestamp.value = 7 * 24 * 60 * 60 * 1000 // 1 week in milliseconds
      break
  }
}

// Computed property to get selected plan details
const selectedPlan = computed(() => {
  return plans.value.find(plan => plan.id === selectPlanName.value)
})

// Check if user details are pre-filled (for UI state)
const isUsernamePrefilled = computed(() => {
  return parsedUserDetails?.username && parsedUserDetails.username.trim() !== ''
})

const isEmailPrefilled = computed(() => {
  return parsedUserDetails?.email && parsedUserDetails.email.trim() !== ''
})

function selectPlan(planId: string) {
  selectPlanName.value = planId as 'student' | 'pro' | 'hobbyist'
  plans.value.forEach((plan) => {
    plan.popular = plan.id === planId;
  });
  setExpiry(planId) // set expiry whenever plan changes
}

function proceedToCheckout(planId: string) {
  selectPlanName.value = planId as 'student' | 'pro' | 'hobbyist'
  setExpiry(planId) // Ensure expiry is set when proceeding to checkout
  showCheckout.value = true
  // Update URL without navigation
  router.replace({ name: 'upgrade', params: { plan: planId } })
}

function goBackToPlans() {
  showCheckout.value = false
  // Reset only phone field when going back (keep user details if they were pre-filled)
  paymentForm.value = {
    username: parsedUserDetails?.username || '',
    email: parsedUserDetails?.email || '',
    phone: '',
  }
  router.replace({ name: 'upgrade' })
}

// Handle M-Pesa payment
async function handlePayment() {
  if (!isFormValid.value) {
    toast.error('Please fill in all required fields with valid information.', { duration: 5000 })
    return
  }
  
  isProcessing.value = true
  
  try {
    // Get the actual expiry timestamp for submission
    const actualExpiryTimestamp = now.value + (expiryTimestamp.value || 0)
    
    // Simulate M-Pesa payment API call
    const paymentData = {
      plan: selectPlanName.value,
      planName: selectedPlan.value?.name,
      amount: selectedPlan.value?.price,
      duration: selectedPlan.value?.duration,
      phone: paymentForm.value.phone,
      email: paymentForm.value.email,
      username: paymentForm.value.username,
      expiryTimestamp: actualExpiryTimestamp,
      expireDuration: expiryTimestamp.value, // Duration in milliseconds
      price: `${selectedPlan.value?.price} Ksh`
    }
    
    console.log('Initiating M-Pesa payment:', paymentData)
    
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 3000))
    
    // Show success message
    alert(`M-Pesa prompt sent to ${paymentForm.value.phone}. Please check your phone and enter your M-Pesa PIN to complete the payment.`)
    
    // Simulate waiting for M-Pesa callback
    setTimeout(() => {
      alert(`Payment successful! Your ${selectedPlan.value?.name} plan is now active and will expire on ${expiryDate.value}.`)
      // In a real app, you would redirect to dashboard or success page
      // router.push('/dashboard')
    }, 5000)
    
  } catch (error) {
    console.error('Payment failed:', error)
    alert('Payment failed. Please try again.')
  } finally {
    isProcessing.value = false
  }
}

watch(selectPlanName, (newVal) => {
  if (!showCheckout.value) {
    router.replace({ name: 'upgrade', params: { plan: newVal } })
  }
  setExpiry(newVal) // Update expiry when plan changes
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})

onMounted(() => {
  // Start ticking every second for reactive time updates
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)

  // Pre-fill form data from user details
  paymentForm.value = {
    username: parsedUserDetails?.username || '',
    email: parsedUserDetails?.email || '',
    phone: parsedUserDetails?.phone || ''
  }
  
  if (planName) {
    // Find and select the plan from URL
    const foundPlan = plans.value.find(p => p.id === planName)
    if (foundPlan) {
      selectPlan(planName)
      showCheckout.value = true
    }
  } else {
    // Set initial expiry for default plan
    setExpiry(selectPlanName.value)
  }
})
</script>

<template>
  <div class="min-h-screen py-6 px-4 sm:px-6 lg:px-8">
    <!-- Back Button -->
    <div class="flex w-full mb-6">
      <button 
        @click="showCheckout ? goBackToPlans() : $router.back()" 
        class="text-gray-600 hover:bg-gray-400 rounded-md hover:text-white w-[35px] h-[35px] flex items-center justify-center transition-colors duration-200"
        :title="showCheckout ? 'Back to Plans' : 'Go Back'"
      >
        <i class="pi pi-arrow-left text-lg font-semibold"></i>
      </button>
    </div>

    <!-- Plans Selection View -->
    <div v-if="!showCheckout">
      <div class="max-w-7xl mx-auto text-center mb-12">
        <h1 class="text-3xl font-bold text-gray-900 sm:text-4xl">
          Choose Your Plan
        </h1>
        <p class="mt-4 text-gray-600 text-lg">
          Flexible pricing designed for Students, Professionals, and Hobbyists.
        </p>
      </div>
  
      <div class="grid gap-8 md:grid-cols-3 max-w-7xl mx-auto">
        <div 
          v-for="plan in plans" 
          :key="plan.id"
          class="relative flex flex-col bg-white border rounded-2xl shadow-sm hover:shadow-lg transition-all duration-200 overflow-hidden cursor-pointer"
          :class="plan.popular ? 'border-blue-600 ring-2 ring-blue-600 transform scale-105' : 'border-gray-200'"
          @click="selectPlan(plan.id)"
        >
          <!-- SELECTED Badge -->
          <div v-if="plan.popular" class="absolute top-0 right-0 bg-blue-600 text-white px-3 py-1 text-xs font-semibold rounded-bl-lg">
            SELECTED
          </div>

          <div class="p-6 flex-grow flex flex-col">
            <h2 class="text-xl font-semibold text-gray-900 mb-2">
              {{ plan.name }}
            </h2>
            <p class="text-gray-600 mb-4 text-sm leading-relaxed">{{ plan.description }}</p>
  
            <div class="mb-6">
              <span class="text-3xl font-bold text-gray-900">{{ plan.price }} Ksh</span>
              <span class="text-gray-500 text-sm ml-1">{{ plan.duration }}</span>
            </div>
  
            <ul class="space-y-3 flex-grow mb-6">
              <li v-for="feature in plan.features" :key="feature" class="flex items-start text-gray-700 text-sm">
                <i class="pi pi-check text-green-600 mr-3 mt-0.5 text-xs"></i>
                <span class="leading-relaxed">{{ feature }}</span>
              </li>
            </ul>
  
            <button 
              @click.stop="proceedToCheckout(plan.id)" 
              class="w-full py-3 px-4 rounded-lg text-white font-medium transition-all duration-200 transform hover:scale-105 active:scale-95"
              :class="plan.popular
                ? 'bg-blue-600 hover:bg-blue-700 shadow-md'
                : 'bg-gray-800 hover:bg-gray-900 shadow-md'
              "
            >
              Get {{ plan.name }} Plan
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Checkout View -->
    <div v-else class="max-w-2xl mx-auto">
      <div class="bg-white rounded-2xl shadow-lg overflow-hidden">
        <!-- Header -->
        <div class="bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-8 text-white">
          <h2 class="text-2xl font-bold mb-2">Complete Your Purchase</h2>
          <p class="text-blue-100">You're upgrading to the {{ selectedPlan?.name }} plan</p>
        </div>

        <!-- Plan Summary -->
        <div class="p-6 border-b border-gray-200">
          <div class="flex justify-between items-center mb-4">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">{{ selectedPlan?.name }} Plan</h3>
              <p class="text-gray-600 text-sm">{{ selectedPlan?.description }}</p>
            </div>
            <div class="text-right">
              <div class="text-2xl font-bold text-gray-900">{{ selectedPlan?.price }} Ksh</div>
              <div class="text-gray-500 text-sm">{{ selectedPlan?.duration }}</div>
            </div>
          </div>

          <!-- Features Summary -->
          <div class="bg-gray-50 rounded-lg p-4">
            <h4 class="font-medium text-gray-900 mb-3">What's included:</h4>
            <div class="grid grid-cols-1 gap-2">
              <div v-for="feature in selectedPlan?.features" :key="feature" class="flex items-center text-sm text-gray-700">
                <i class="pi pi-check text-green-600 mr-2"></i>
                <span>{{ feature }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Payment Form -->
        <div class="p-6">
          <h4 class="font-medium text-gray-900 mb-4">Payment Information</h4>
          
          <form @submit.prevent="handlePayment" class="space-y-4">
            <!-- User Information -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label for="username" class="block text-sm font-medium text-gray-700 mb-1">
                  Username
                </label>
                <input
                  id="username"
                  v-model="paymentForm.username"
                  type="text"
                  required
                  :disabled="isUsernamePrefilled"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
                  :class="isUsernamePrefilled ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'"
                  placeholder="Enter your username"
                />
                <p v-if="isUsernamePrefilled" class="text-xs text-gray-500 mt-1">
                  Using your account username
                </p>
              </div>
              
              <div>
                <label for="email" class="block text-sm font-medium text-gray-700 mb-1">
                  Email Address
                </label>
                <input
                  id="email"
                  v-model="paymentForm.email"
                  type="email"
                  required
                  :disabled="isEmailPrefilled"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
                  :class="isEmailPrefilled ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'"
                  placeholder="your.email@example.com"
                />
                <p v-if="isEmailPrefilled" class="text-xs text-gray-500 mt-1">
                  Using your account email
                </p>
              </div>
            </div>

            <div>
              <label for="phone" class="block text-sm font-medium text-gray-700 mb-1">
                M-Pesa Phone Number
              </label>
              <input
                id="phone"
                v-model="paymentForm.phone"
                type="tel"
                required
                pattern="^(\+254|0)[17][0-9]{8}$"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
                placeholder="0712345678 or +254712345678"
              />
              <p class="text-xs text-gray-500 mt-1">
                Enter your Safaricom M-Pesa number
              </p>
            </div>

            <!-- Plan Duration & Expiry -->
            <div class="bg-blue-50 rounded-lg p-4 border border-blue-200">
              <div class="flex items-center justify-between mb-2">
                <div>
                  <h5 class="font-medium text-blue-900 mb-1">Plan Duration</h5>
                  <p class="text-sm text-blue-700">
                    Your {{ selectedPlan?.name }} plan will be active {{ selectedPlan?.duration }}
                  </p>
                </div>
                <div class="text-right">
                  <p class="text-xs text-blue-600 uppercase font-semibold">Expires</p>
                  <p class="text-sm font-medium text-blue-900">{{ expiryDate }}</p>
                </div>
              </div>
              <div v-if="timeLeft && timeLeft !== 'Expired'" class="text-center">
                <div class="inline-flex items-center bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm font-medium">
                  <i class="pi pi-clock mr-1"></i>
                  {{ timeLeft }} remaining after purchase
                </div>
              </div>
            </div>

            <!-- M-Pesa Payment Method -->
            <div class="border border-green-200 rounded-lg p-4 bg-green-50">
              <div class="flex items-center mb-3">
                <div class="w-8 h-8 bg-green-600 rounded-full flex items-center justify-center mr-3">
                  <i class="pi pi-mobile text-white text-sm"></i>
                </div>
                <div>
                  <h5 class="font-medium text-green-900">M-Pesa Payment</h5>
                  <p class="text-sm text-green-700">Safe and secure mobile payment</p>
                </div>
              </div>
              <div class="text-sm text-green-800">
                <p class="mb-1">• You'll receive an M-Pesa prompt on your phone</p>
                <p class="mb-1">• Enter your M-Pesa PIN to complete payment</p>
                <p>• Payment confirmation will be sent via SMS</p>
              </div>
            </div>

            <!-- Form Validation Message -->
            <div v-if="!isFormValid && (paymentForm.phone || paymentForm.username || paymentForm.email)" 
                 class="bg-red-50 border border-red-200 rounded-lg p-3">
              <div class="flex items-center">
                <i class="pi pi-exclamation-triangle text-red-600 mr-2"></i>
                <p class="text-sm text-red-800">
                  Please ensure all fields are filled correctly:
                </p>
              </div>
              <ul class="text-xs text-red-700 mt-2 ml-6">
                <li v-if="!paymentForm.username.trim()">Username is required</li>
                <li v-if="!paymentForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(paymentForm.email)">Valid email is required</li>
                <li v-if="!/^(\+254|0)[17][0-9]{8}$/.test(paymentForm.phone)">Valid Kenyan phone number is required</li>
              </ul>
            </div>

            <!-- Action Buttons -->
            <div class="flex gap-4 pt-4">
              <button 
                @click="goBackToPlans"
                type="button"
                class="flex-1 py-3 px-4 border border-gray-300 rounded-lg text-gray-700 font-medium hover:bg-gray-50 transition-colors duration-200"
              >
                Back to Plans
              </button>
              <button 
                type="submit"
                :disabled="!isFormValid || isProcessing"
                class="flex-1 py-3 px-4 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors duration-200 flex items-center justify-center"
              >
                <i v-if="isProcessing" class="pi pi-spinner pi-spin mr-2"></i>
                {{ isProcessing ? 'Processing...' : `Pay ${selectedPlan?.price} Ksh via M-Pesa` }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
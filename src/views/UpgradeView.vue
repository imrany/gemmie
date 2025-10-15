<script lang="ts" setup>
import { API_BASE_URL, getTransaction, plans } from '@/utils/globals';
import type { Ref } from 'vue';
import { onMounted, ref, watch, computed, inject, onUnmounted, reactive } from 'vue';
import { useRoute } from 'vue-router';
import { useRouter } from 'vue-router';
import { toast } from 'vue-sonner';

const route = useRoute()
const router = useRouter()
const planName = route.params.plan as 'student' | 'pro' | 'hobbyist' | undefined
const selectPlanName = ref<'student' | 'pro' | 'hobbyist'>(planName ?? 'pro')
const globalState = inject("globalState") as {
  parsedUserDetails: Ref<any>,
}
const parsedUserDetails = globalState.parsedUserDetails
const showCheckout = ref(false)
const isProcessing = ref(false)

const planDuration = ref<number | null>(null) // duration in ms
const expiryTimestamp = ref<number | null>(null) // absolute expiry time in ms
const now = ref(Date.now())
let timer: number | null = null
let paymentCheckInterval: number | null = null

// Expiry date (absolute)
const expiryDate = computed(() => {
  if (!expiryTimestamp.value) return ''
  return new Date(expiryTimestamp.value).toLocaleString('en-KE', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
})

// Current plan time left (for existing active plan)
const currentPlanTimeLeft = computed(() => {
  if (!parsedUserDetails?.value.expiry_timestamp) return ''
  const expiryMs =
    parsedUserDetails.value.expiry_timestamp < 1e12
      ? parsedUserDetails.value.expiry_timestamp * 1000
      : parsedUserDetails.value.expiry_timestamp

  const diff = expiryMs - now.value
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

// Plan duration (local only)
function setDuration(planId: string) {
  switch (planId) {
    case 'student':
      planDuration.value = 5 * 60 * 60 * 1000 // 5h
      break
    case 'pro':
      planDuration.value = 24 * 60 * 60 * 1000 // 24h
      break
    case 'hobbyist':
      planDuration.value = 7 * 24 * 60 * 60 * 1000 // 1 week
      break
  }
  expiryTimestamp.value = now.value + (planDuration.value || 0)
}

// Payment form data - create local reactive references
const paymentForm = reactive({
  username: parsedUserDetails.value?.username || '',
  email: parsedUserDetails.value?.email || '',
  phone: '',
})

// Form validation
const isFormValid = computed(() => {
  const phoneRegex = /^(\+254|0)[17][0-9]{8}$/
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  // Check username
  const hasValidUsername = paymentForm.username.trim() !== ''

  // Check email
  const hasValidEmail = emailRegex.test(paymentForm.email)

  // Check phone
  const hasValidPhone = phoneRegex.test(paymentForm.phone)

  return hasValidUsername && hasValidEmail && hasValidPhone
})

// Computed property to get selected plan details
const selectedPlan = computed(() => {
  return plans.value.find(plan => plan.id === selectPlanName.value)
})

// Check if user details are pre-filled (for UI state)
const isUsernamePrefilled = computed(() => {
  return parsedUserDetails.value?.username && parsedUserDetails.value?.username.trim() !== ''
})

const isEmailPrefilled = computed(() => {
  return parsedUserDetails.value?.email && parsedUserDetails.value?.email.trim() !== ''
})

const isPhonePrefilled = computed(() => {
  return !!(parsedUserDetails.value?.phone_number && parsedUserDetails.value?.phone_number.trim() !== '')
})

// Check if user has an active plan
const hasActivePlan = computed(() => {
  if (!parsedUserDetails.value?.expiry_timestamp) return false
  const expiryMs = parsedUserDetails.value.expiry_timestamp < 1e12
    ? parsedUserDetails.value.expiry_timestamp * 1000
    : parsedUserDetails.value.expiry_timestamp
  return expiryMs > now.value
})

function selectPlan(planId: string) {
  selectPlanName.value = planId as 'student' | 'pro' | 'hobbyist'
  plans.value.forEach((plan) => {
    plan.popular = plan.id === planId;
  });
  setDuration(planId) // set expiry whenever plan changes
}

function proceedToCheckout(planId: string) {
  if(parsedUserDetails.value){
    selectPlanName.value = planId as 'student' | 'pro' | 'hobbyist'
    selectPlan(planId)
    setDuration(planId) // Ensure expiry is set when proceeding to checkout
    showCheckout.value = true
    // Update URL without navigation
    router.replace({ name: 'upgrade', params: { plan: planId } })
    return
  }
  router.push("/?from=upgrade")
}

function goBackToPlans() {
  showCheckout.value = false
  // Reset only phone field when going back (keep user details if they were pre-filled)
  paymentForm.phone = ''
  router.replace({ name: 'upgrade' })
}

// Clean up intervals
function clearIntervals() {
  if (paymentCheckInterval) {
    clearInterval(paymentCheckInterval)
    paymentCheckInterval = null
  }
}

// Handle M-Pesa payment
async function handlePayment() {
  if (!isFormValid.value) {
    toast.error('Please fill in all required fields with valid information.', { duration: 5000 })
    return
  }

  isProcessing.value = true
  clearIntervals() // Clear any existing intervals

  try {
    // Get the actual expiry timestamp for submission
    const actualExpiryTimestamp = now.value + (planDuration.value || 0)

    const paymentData = {
      expiry_timestamp: actualExpiryTimestamp,
      expire_duration: planDuration.value,
      external_reference: `${parsedUserDetails.value?.username}-${Date.now()}`,
      plan: selectPlanName.value,
      plan_name: selectedPlan.value?.name,
      amount: selectedPlan.value?.price,
      duration: selectedPlan.value?.duration,
      phone_number: paymentForm.phone,
      email: paymentForm.email,
      username: paymentForm.username,
      price: `${selectedPlan.value?.price} Ksh`
    }

    console.log('Initiating M-Pesa payment')
    const stkResults = await sendSTK(paymentData)
    
    if (!stkResults || !stkResults.data || stkResults.data.success === false) {
      isProcessing.value = false
      toast.error('Failed to initiate payment. Please try again.', { duration: 5000 })
      return
    }

    console.log('STK Push sent successfully:', stkResults)
    toast.success(`M-Pesa prompt sent to ${paymentForm.phone}. Please check your phone and enter your M-Pesa PIN to complete the payment.`, { duration: 8000 })
    
    // Store external reference safely
    const externalRef = stkResults.data.external_reference
    if (externalRef) {
      localStorage.setItem("external_reference", JSON.stringify(externalRef))
    }

    // Payment confirmation logic with better error handling
    let attempts = 0;
    const maxAttempts = 10; // Check for up to 10 attempts (about 3.5 minutes)
    
    paymentCheckInterval = window.setInterval(async () => {
      attempts++;
      console.log(`Payment check attempt ${attempts}/${maxAttempts}`)
      
      try {
        const ext = JSON.parse(localStorage.getItem("external_reference") || '""')
        if (!ext) {
          console.error('No external reference found')
          clearIntervals()
          isProcessing.value = false
          toast.error("Payment reference lost. Please try again.", { duration: 6000 })
          return
        }

        const transResults = await getTransaction(ext)
        console.log('Transaction check result:', transResults)
        
        // Check for successful payment
        if (transResults?.data?.Status === "Success" || transResults?.success === true) {
          clearIntervals()
          // localStorage.removeItem("external_reference") // Clean up
          
          toast.success(`Payment successful! Your ${selectedPlan.value?.name} plan is now active until ${expiryDate.value}.`, { 
            duration: 10000,
            important: true
          })
          
          isProcessing.value = false
          
          // Navigate back to home after a short delay
          setTimeout(() => {
            router.push('/')
          }, 2000)
          return
        }
        
        // Check for failed payment
        if (transResults?.data?.Status === "Failed" || transResults?.success === false) {
          clearIntervals()
          localStorage.removeItem("external_reference")
          
          toast.error("Payment failed. Please try again or contact support.", { duration: 8000 })
          isProcessing.value = false
          return
        }
        
        // Still pending - show progress
        if (attempts <= 3) {
          toast.info(`Waiting for payment confirmation... (${attempts}/${maxAttempts})`, { 
            duration: 3000,
            description: "Please complete the payment on your phone"
          })
        }
        
      } catch (error) {
        console.error('Error checking transaction:', error)
      }
      
      // Max attempts reached
      if (attempts >= maxAttempts) { 
        clearIntervals()
        toast.error("Payment confirmation timeout. If you completed the payment, please refresh this page or contact support.", { 
          duration: 10000,
          action: {
            label: "Refresh Page",
            onClick: () => window.location.reload()
          }
        })
        isProcessing.value = false
      }
    }, 20000) // Check every 20 seconds

  } catch (error) {
    console.error('Payment error:', error)
    clearIntervals()
    isProcessing.value = false
    toast.error('Payment failed. Please try again.', {
      duration: 5000,
      action: {
        label: "Try again",
        onClick: () => handlePayment()
      }
    })
  }
}

async function sendSTK(paymentData: any) {
  try {
    const res = await fetch(`${API_BASE_URL}/payments/stk`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        external_reference: paymentData.external_reference,
        phone_number: paymentData.phone_number,
        amount: paymentData.amount,
      }),
    });
    
    if (!res.ok) {
      throw new Error(`HTTP error! status: ${res.status}`)
    }
    
    return await res.json();
  } catch (error) {
    console.error('STK Push error:', error)
    toast.error("Failed to send M-Pesa prompt. Please check your connection and try again.", { duration: 5000 })
    throw error
  }
}

function handleBack(){
  if(window.history.state.back){
    router.back()
    return
  }
  router.push("/")
}

watch(selectPlanName, (newVal) => {
  if (!showCheckout.value) {
    router.replace({ name: 'upgrade', params: { plan: newVal } })
  }
  setDuration(newVal) // Update expiry when plan changes
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  clearIntervals()
})

onMounted(() => {
  if(!parsedUserDetails?.value){
    return
  }
  // Check if user has an active plan
  const expiry = parsedUserDetails?.value.expiry_timestamp
  if (expiry && expiry * 1000 > Date.now()) {
    toast.info(`You currently have an active ${parsedUserDetails.value.plan_name} plan.`, {
      duration: Infinity,
      description: `Time remaining: ${currentPlanTimeLeft.value}`,
      action: {
        label: "Go Back",
        onClick: () => router.back()
      }
    })
  }

  // Keep ticking every second
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)

  // Pre-fill user info
  paymentForm.username = parsedUserDetails.value?.username || ''
  paymentForm.email = parsedUserDetails.value?.email || ''
  paymentForm.phone = parsedUserDetails.value?.phone_number || ''

  // Select plan from URL or default
  if (planName) {
    selectPlan(planName)
    showCheckout.value = true
  } else {
    setDuration(selectPlanName.value)
  }
})

</script>

<template>
  <div class="min-h-screen py-6 px-4 sm:px-6 lg:px-8 bg-gray-50 dark:bg-gray-900 transition-colors duration-300">
    <!-- Back Button -->
    <div class="flex w-full mb-6">
      <button @click="showCheckout ? goBackToPlans() : handleBack()"
        class="text-gray-600 dark:text-gray-400 hover:bg-gray-400 dark:hover:bg-gray-600 rounded-md hover:text-white w-[35px] h-[35px] flex items-center justify-center transition-colors duration-200"
        :title="showCheckout ? 'Back to Plans' : 'Go Back'">
        <i class="pi pi-arrow-left text-lg font-semibold"></i>
      </button>
    </div>

    <!-- Active Plan Notice -->
    <div v-if="hasActivePlan && !showCheckout" class="max-w-7xl mx-auto mb-8">
      <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 transition-colors duration-200">
        <div class="flex items-center">
          <i class="pi pi-info-circle text-blue-600 dark:text-blue-400 mr-3"></i>
          <div>
            <h3 class="text-blue-900 dark:text-blue-300 font-medium">Current Active Plan</h3>
            <p class="text-blue-700 dark:text-blue-400 text-sm">
              You have an active {{ parsedUserDetails.plan_name }} plan with {{ currentPlanTimeLeft }} remaining.
              Purchasing a new plan will replace your current one.
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Plans Selection View -->
    <div v-if="!showCheckout">
      <div class="max-w-7xl mx-auto text-center mb-12">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white sm:text-4xl">
          Choose Your Plan
        </h1>
        <p class="mt-4 text-gray-600 dark:text-gray-400 text-lg">
          Flexible pricing designed for Students, Professionals, and Hobbyists.
        </p>
      </div>

      <div class="grid gap-8 md:grid-cols-3 max-w-7xl mx-auto">
        <div v-for="plan in plans" :key="plan.id"
          class="relative flex flex-col bg-white dark:bg-gray-800 border rounded-2xl shadow-sm hover:shadow-lg transition-all duration-200 overflow-hidden cursor-pointer"
          :class="plan.popular 
            ? 'border-blue-600 dark:border-blue-500 ring-2 ring-blue-600 dark:ring-blue-500 transform scale-105' 
            : 'border-gray-200 dark:border-gray-700'"
          @click="selectPlan(plan.id)">
          <!-- SELECTED Badge -->
          <div v-if="plan.popular"
            class="absolute top-0 right-0 bg-blue-600 dark:bg-blue-500 text-white px-3 py-1 text-xs font-semibold rounded-bl-lg">
            SELECTED
          </div>

          <!-- Current Plan Badge -->
          <div v-if="hasActivePlan && parsedUserDetails.plan_name?.toLowerCase().includes(plan.name.toLowerCase())"
            class="absolute top-0 left-0 bg-green-600 dark:bg-green-500 text-white px-3 py-1 text-xs font-semibold rounded-br-lg">
            CURRENT
          </div>

          <div class="p-6 flex-grow flex flex-col">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
              {{ plan.name }}
            </h2>
            <p class="text-gray-600 dark:text-gray-400 mb-4 text-sm leading-relaxed">{{ plan.description }}</p>

            <div class="mb-6">
              <span class="text-3xl font-bold text-gray-900 dark:text-white">{{ plan.price }} Ksh</span>
              <span class="text-gray-500 dark:text-gray-400 text-sm ml-1">{{ plan.duration }}</span>
            </div>

            <ul class="space-y-3 flex-grow mb-6">
              <li v-for="feature in plan.features" :key="feature" class="flex items-start text-gray-700 dark:text-gray-300 text-sm">
                <i class="pi pi-check text-green-600 dark:text-green-400 mr-3 mt-0.5 text-xs"></i>
                <span class="leading-relaxed">{{ feature }}</span>
              </li>
            </ul>

            <button @click.stop="proceedToCheckout(plan.id)"
              class="w-full py-3 px-4 rounded-lg text-white font-medium transition-all duration-200 transform hover:scale-105 active:scale-95"
              :class="plan.popular
                ? 'bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 shadow-md'
                : 'bg-gray-800 hover:bg-gray-900 dark:bg-gray-700 dark:hover:bg-gray-600 shadow-md'
                ">
              {{ hasActivePlan && parsedUserDetails.plan_name?.toLowerCase().includes(plan.name.toLowerCase()) 
                ? `Renew ${plan.name} Plan` 
                : `Get ${plan.name} Plan` }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Checkout View -->
    <div v-else-if="parsedUserDetails&&showCheckout" class="max-w-2xl mx-auto">
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg overflow-hidden transition-colors duration-300">
        <!-- Header -->
        <div class="bg-gradient-to-r from-blue-600 to-blue-700 dark:from-blue-700 dark:to-blue-800 px-6 py-8 text-white">
          <h2 class="text-2xl font-bold mb-2">Complete Your Purchase</h2>
          <p class="text-blue-100 dark:text-blue-200">
            {{ hasActivePlan ? `Upgrading to the ${selectedPlan?.name} plan` : `You're purchasing the ${selectedPlan?.name} plan` }}
          </p>
        </div>

        <!-- Current Plan Notice in Checkout -->
        <div v-if="hasActivePlan" class="bg-yellow-50 dark:bg-yellow-900/20 border-b border-yellow-200 dark:border-yellow-800 px-6 py-4 transition-colors duration-200">
          <div class="flex items-center">
            <i class="pi pi-exclamation-triangle text-yellow-600 dark:text-yellow-400 mr-3"></i>
            <div>
              <p class="text-yellow-800 dark:text-yellow-300 text-sm">
                <strong>Note:</strong> You currently have {{ currentPlanTimeLeft }} remaining on your {{ parsedUserDetails.plan_name }} plan. 
                This purchase will replace your current plan immediately.
              </p>
            </div>
          </div>
        </div>

        <!-- Plan Summary -->
        <div class="p-6 border-b border-gray-200 dark:border-gray-700 transition-colors duration-200">
          <div class="flex justify-between items-center mb-4">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ selectedPlan?.name }} Plan</h3>
              <p class="text-gray-600 dark:text-gray-400 text-sm">{{ selectedPlan?.description }}</p>
            </div>
            <div class="text-right">
              <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ selectedPlan?.price }} Ksh</div>
              <div class="text-gray-500 dark:text-gray-400 text-sm">{{ selectedPlan?.duration }}</div>
            </div>
          </div>

          <!-- Features Summary -->
          <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 transition-colors duration-200">
            <h4 class="font-medium text-gray-900 dark:text-white mb-3">What's included:</h4>
            <div class="grid grid-cols-1 gap-2">
              <div v-for="feature in selectedPlan?.features" :key="feature"
                class="flex items-center text-sm text-gray-700 dark:text-gray-300">
                <i class="pi pi-check text-green-600 dark:text-green-400 mr-2"></i>
                <span>{{ feature }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Payment Form -->
        <div class="p-6">
          <h4 class="font-medium text-gray-900 dark:text-white mb-4">Payment Information</h4>

          <form @submit.prevent="handlePayment" class="space-y-4">
            <!-- User Information -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Username
                </label>
                <input id="username" v-model="paymentForm.username" type="text" required :disabled="isUsernamePrefilled"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                  :class="isUsernamePrefilled ? 'bg-gray-100 dark:bg-gray-600 cursor-not-allowed' : 'bg-white dark:bg-gray-700'"
                  placeholder="Enter your username" />
                <p v-if="isUsernamePrefilled" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  Using your account username
                </p>
              </div>

              <div>
                <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Email Address
                </label>
                <input id="email" v-model="paymentForm.email" type="email" required :disabled="isEmailPrefilled"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                  :class="isEmailPrefilled ? 'bg-gray-100 dark:bg-gray-600 cursor-not-allowed' : 'bg-white dark:bg-gray-700'"
                  placeholder="your.email@example.com" />
                <p v-if="isEmailPrefilled" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  Using your account email
                </p>
              </div>
            </div>

            <div>
              <label for="phone" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                M-Pesa Phone Number
              </label>
              <input id="phone" v-model="paymentForm.phone" type="tel" required
                pattern="^(\+254|0)[17][0-9]{8}$"
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                placeholder="0712345678 or +254712345678" />
              <div class="flex items-center gap-2 mt-1" v-if="isPhonePrefilled && !paymentForm.phone">
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  Using your account phone number: {{ parsedUserDetails.phone_number }}
                </p>
                <button type="button" class="text-blue-600 dark:text-blue-400 hover:underline text-xs" @click="paymentForm.phone = parsedUserDetails.phone_number">
                  Use Account Phone
                </button>
              </div>
              <p v-if="!isPhonePrefilled" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Enter your Safaricom M-Pesa number
              </p>
            </div>

            <!-- Plan Duration & Expiry -->
            <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4 border border-blue-200 dark:border-blue-800 transition-colors duration-200">
              <div class="flex items-center justify-between mb-2">
                <div>
                  <h5 class="font-medium text-blue-900 dark:text-blue-300 mb-1">Plan Duration</h5>
                  <p class="text-sm text-blue-700 dark:text-blue-400">
                    Your {{ selectedPlan?.name }} plan will be active for {{ selectedPlan?.duration }}
                  </p>
                </div>
                <div class="text-right">
                  <p class="text-xs text-blue-600 dark:text-blue-400 uppercase font-semibold">Expires</p>
                  <p class="text-sm font-medium text-blue-900 dark:text-blue-300">{{ expiryDate }}</p>
                </div>
              </div>
            </div>

            <!-- M-Pesa Payment Method -->
            <div class="border border-green-200 dark:border-green-800 rounded-lg p-4 bg-green-50 dark:bg-green-900/20 transition-colors duration-200">
              <div class="flex items-center mb-3">
                <div class="w-8 h-8 bg-green-600 dark:bg-green-500 rounded-full flex items-center justify-center mr-3">
                  <i class="pi pi-mobile text-white text-sm"></i>
                </div>
                <div>
                  <h5 class="font-medium text-green-900 dark:text-green-300">M-Pesa Payment</h5>
                  <p class="text-sm text-green-700 dark:text-green-400">Safe and secure mobile payment</p>
                </div>
              </div>
              <div class="text-sm text-green-800 dark:text-green-300">
                <p class="mb-1">• You'll receive an M-Pesa prompt on your phone</p>
                <p class="mb-1">• Enter your M-Pesa PIN to complete payment</p>
                <p>• Payment confirmation will be automatic</p>
              </div>
            </div>

            <!-- Form Validation Message -->
            <div v-if="!isFormValid && (paymentForm.phone || paymentForm.username || paymentForm.email)" 
                 class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3 transition-colors duration-200">
              <div class="flex items-center">
                <i class="pi pi-exclamation-triangle text-red-600 dark:text-red-400 mr-2"></i>
                <p class="text-sm text-red-800 dark:text-red-300">
                  Please ensure all fields are filled correctly:
                </p>
              </div>
              <ul class="text-xs text-red-700 dark:text-red-400 mt-2 ml-6">
                <li v-if="!paymentForm.username.trim()">Username is required</li>
                <li v-if="!paymentForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(paymentForm.email)">Valid email is required</li>
                <li v-if="!/^(\+254|0)[17][0-9]{8}$/.test(paymentForm.phone)">Valid Kenyan phone number is required</li>
              </ul>
            </div>

            <!-- Action Buttons -->
            <div class="flex gap-4 pt-4">
              <button @click="goBackToPlans" type="button" :disabled="isProcessing"
                class="flex-1 py-3 px-4 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-300 font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed">
                Back to Plans
              </button>
              <button type="submit" :disabled="!isFormValid || isProcessing"
                class="flex-1 py-3 px-4 bg-green-600 dark:bg-green-500 text-white rounded-lg font-medium hover:bg-green-700 dark:hover:bg-green-600 disabled:bg-gray-400 dark:disabled:bg-gray-600 disabled:cursor-not-allowed transition-colors duration-200 flex items-center justify-center">
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
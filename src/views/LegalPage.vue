<script setup lang="ts">
import { ref } from "vue";
import { ChevronLeft } from "lucide-vue-next";
import { useRouter, useRoute } from "vue-router";

const router = useRouter();
const route = useRoute();

// Determine active tab from route params or default to privacy
const activeTab = ref<"privacy" | "terms">(
    (route.params.section as "privacy" | "terms") || "privacy",
);

// Update route when tab changes
function switchTab(tab: "privacy" | "terms") {
    activeTab.value = tab;
    router.push({ name: "legal", params: { section: tab } });
}

const lastUpdated = "October 27, 2025";
</script>

<template>
    <div
        class="min-h-screen font-light bg-gray-50 dark:bg-gray-900 py-8 px-4 sm:px-6 lg:px-8 transition-colors duration-300"
    >
        <div class="max-w-4xl mx-auto">
            <!-- Header -->
            <div class="text-center mb-8">
                <p class="text-gray-600 dark:text-gray-400">
                    Last updated: {{ lastUpdated }}
                </p>
            </div>

            <!-- Tab Navigation -->
            <div
                class="bg-white dark:bg-gray-800 rounded-lg shadow-sm mb-6 p-1 flex gap-2 transition-colors duration-300"
            >
                <button
                    @click="switchTab('privacy')"
                    :class="
                        activeTab === 'privacy'
                            ? 'bg-blue-600 text-white'
                            : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                    "
                    class="flex-1 py-2 px-4 text-sm rounded-md font-medium transition-all duration-200"
                >
                    Privacy Policy
                </button>
                <button
                    @click="switchTab('terms')"
                    :class="
                        activeTab === 'terms'
                            ? 'bg-blue-600 text-white'
                            : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                    "
                    class="flex-1 py-2 px-4 text-sm rounded-md font-medium transition-all duration-200"
                >
                    Terms of Service
                </button>
            </div>

            <!-- Privacy Policy Content -->
            <div
                v-if="activeTab === 'privacy'"
                class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 md:p-8 transition-colors duration-300"
            >
                <div class="prose prose-gray dark:prose-invert max-w-none">
                    <h2
                        class="text-2xl font-bold text-gray-900 dark:text-white mb-4"
                    >
                        Privacy Policy
                    </h2>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            1. Introduction
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            We are committed to protecting your privacy in
                            accordance with the Data Protection Act, 2019
                            (Kenya) and other applicable Kenyan laws. This
                            Privacy Policy explains how we collect, use,
                            disclose, and safeguard your information when you
                            use our AI-powered chat application.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            2. Information We Collect
                        </h3>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            2.1 Personal Information
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-2"
                        >
                            We collect the following personal information:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                Username and email address (for account
                                creation)
                            </li>
                            <li>
                                M-Pesa phone number (for payment processing)
                            </li>
                            <li>Chat history and conversations</li>
                            <li>User preferences and settings</li>
                            <li>Device information and IP address</li>
                        </ul>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            2.2 Automatically Collected Information
                        </h4>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>Browser type and version</li>
                            <li>Operating system</li>
                            <li>Usage patterns and interaction data</li>
                            <li>Error logs and diagnostic data</li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            3. How We Use Your Information
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-2"
                        >
                            We use your information for:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                Providing and maintaining the Gemmie service
                            </li>
                            <li>Processing payments via M-Pesa</li>
                            <li>Personalizing your experience</li>
                            <li>
                                Syncing data across your devices (if enabled)
                            </li>
                            <li>Improving our AI models and services</li>
                            <li>
                                Communicating with you about updates and
                                features
                            </li>
                            <li>Detecting and preventing fraud or abuse</li>
                            <li>Complying with legal obligations</li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            4. Data Storage and Security
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-3"
                        >
                            Your data is stored securely using industry-standard
                            encryption. We implement appropriate technical and
                            organizational measures to protect your personal
                            information against unauthorized access, alteration,
                            disclosure, or destruction.
                        </p>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            <strong>Local Storage:</strong> If sync is disabled,
                            your chat data is stored locally on your device
                            only.
                            <br />
                            <strong>Cloud Storage:</strong> If sync is enabled,
                            your data is encrypted and stored on secure servers.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            5. Data Sharing and Disclosure
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-2"
                        >
                            We do not sell your personal information. We may
                            share your data with:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                <strong>Payment Processors:</strong>
                                M-Pesa for payment processing
                            </li>
                            <li>
                                <strong>AI Service Providers:</strong>
                                Third-party AI services to process your queries
                                (data is anonymized where possible)
                            </li>
                            <li>
                                <strong>Legal Authorities:</strong> When
                                required by Kenyan law or to protect our rights
                            </li>
                            <li>
                                <strong>Service Providers:</strong> Trusted
                                partners who help us operate our service (under
                                strict confidentiality agreements)
                            </li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            6. Your Rights Under Kenyan Law
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-2"
                        >
                            Under the Data Protection Act, 2019, you have the
                            right to:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                <strong>Access:</strong> Request copies of your
                                personal data
                            </li>
                            <li>
                                <strong>Correction:</strong> Request correction
                                of inaccurate data
                            </li>
                            <li>
                                <strong>Deletion:</strong> Request deletion of
                                your data (right to be forgotten)
                            </li>
                            <li>
                                <strong>Data Portability:</strong> Receive your
                                data in a structured format
                            </li>
                            <li>
                                <strong>Withdraw Consent:</strong> Withdraw
                                consent for data processing at any time
                            </li>
                            <li>
                                <strong>Object:</strong> Object to processing of
                                your personal data
                            </li>
                            <li>
                                <strong>Restrict:</strong> Request restriction
                                of processing
                            </li>
                        </ul>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mt-3"
                        >
                            To exercise these rights, contact us at:
                            <a
                                href="mailto:privacy@gemmie.app"
                                class="text-blue-600 dark:text-blue-400 hover:underline"
                                >privacy@gemmie.app</a
                            >
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            7. Data Retention
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            We retain your personal data for as long as your
                            account is active or as needed to provide services.
                            When you delete your account, we will delete or
                            anonymize your data within 30 days, except where
                            retention is required by law.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            8. Children's Privacy
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            Gemmie is not intended for users under 18 years of
                            age. We do not knowingly collect personal
                            information from children under 18. If you believe
                            we have collected information from a minor, please
                            contact us immediately.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            9. Cookies and Tracking
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            We use local storage and session data to maintain
                            your login state and preferences. We do not use
                            third-party tracking cookies for advertising
                            purposes.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            10. International Data Transfers
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            Your data may be processed outside Kenya. We ensure
                            appropriate safeguards are in place to protect your
                            data in accordance with Kenyan data protection laws.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            11. Changes to This Policy
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            We may update this Privacy Policy from time to time.
                            We will notify you of significant changes via email
                            or through the app. Your continued use of Gemmie
                            after changes constitutes acceptance of the updated
                            policy.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            12. Contact Information
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            For questions about this Privacy Policy or to
                            exercise your rights, contact us at:
                        </p>
                        <div
                            class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg mt-3"
                        >
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Email:</strong> privacy@gemmie.app
                            </p>
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Address:</strong> Nairobi, Kenya
                            </p>
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Data Protection Officer:</strong>
                                dpo@gemmie.app
                            </p>
                        </div>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            13. Office of the Data Protection Commissioner
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            You have the right to lodge a complaint with the
                            Office of the Data Protection Commissioner (ODPC) of
                            Kenya:
                        </p>
                        <div
                            class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg mt-3"
                        >
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Website:</strong>
                                <a
                                    href="https://www.odpc.go.ke"
                                    target="_blank"
                                    class="text-blue-600 dark:text-blue-400 hover:underline"
                                    >www.odpc.go.ke</a
                                >
                            </p>
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Email:</strong> info@odpc.go.ke
                            </p>
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Phone:</strong> +254 (020) 2211960
                            </p>
                        </div>
                    </section>
                </div>
            </div>

            <!-- Terms of Service Content -->
            <div
                v-if="activeTab === 'terms'"
                class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 md:p-8 transition-colors duration-300"
            >
                <div class="prose prose-gray dark:prose-invert max-w-none">
                    <h2
                        class="text-2xl font-bold text-gray-900 dark:text-white mb-4"
                    >
                        Terms of Service
                    </h2>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            1. Agreement to Terms
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            By accessing or using Gemmie, you agree to be bound
                            by these Terms of Service and all applicable laws of
                            Kenya. If you do not agree to these terms, do not
                            use our service. These terms constitute a legally
                            binding agreement between you and Gemmie.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            2. Service Description
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            Gemmie is an AI-powered chat application that
                            provides conversational AI services. We offer both
                            free and paid subscription plans. The service is
                            provided "as is" and may be modified, suspended, or
                            discontinued at any time.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            3. Eligibility
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-2"
                        >
                            To use Gemmie, you must:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>Be at least 18 years of age</li>
                            <li>
                                Have the legal capacity to enter into binding
                                contracts under Kenyan law
                            </li>
                            <li>
                                Provide accurate and complete registration
                                information
                            </li>
                            <li>
                                Not be prohibited from using the service under
                                Kenyan or international law
                            </li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            4. Account Registration and Security
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-3"
                        >
                            You are responsible for maintaining the
                            confidentiality of your account credentials. You
                            agree to:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                Provide accurate, current, and complete
                                information
                            </li>
                            <li>Update your information to keep it accurate</li>
                            <li>Keep your password secure and confidential</li>
                            <li>
                                Notify us immediately of any unauthorized access
                            </li>
                            <li>
                                Accept responsibility for all activities under
                                your account
                            </li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            5. Subscription Plans and Payments
                        </h3>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            5.1 Free Plan
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            Free users are limited to 20 requests per 24-hour
                            period. Usage resets daily at midnight EAT (East
                            Africa Time).
                        </p>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            5.2 Paid Plans
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-2"
                        >
                            Paid plans offer unlimited requests and additional
                            features. Payments are processed via M-Pesa:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>All prices are in Kenyan Shillings (KES)</li>
                            <li>
                                Payments are non-refundable unless required by
                                Kenyan consumer protection law
                            </li>
                            <li>
                                Plans auto-renew unless cancelled before the
                                renewal date
                            </li>
                            <li>
                                We reserve the right to change prices with 30
                                days notice
                            </li>
                        </ul>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            5.3 Refund Policy
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            In accordance with the Consumer Protection Act, 2012
                            (Kenya), refunds may be issued if:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                The service is materially different from what
                                was advertised
                            </li>
                            <li>
                                Technical issues prevent you from using the
                                service for more than 48 consecutive hours
                            </li>
                            <li>
                                Request is made within 7 days of purchase with
                                valid reason
                            </li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            6. Acceptable Use Policy
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-3"
                        >
                            You agree NOT to:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                Use the service for any illegal purpose under
                                Kenyan law
                            </li>
                            <li>
                                Violate the Computer Misuse and Cybercrimes Act,
                                2018
                            </li>
                            <li>
                                Generate or distribute harmful, hateful, or
                                discriminatory content
                            </li>
                            <li>
                                Attempt to bypass usage limits or access
                                restrictions
                            </li>
                            <li>
                                Reverse engineer, decompile, or disassemble the
                                service
                            </li>
                            <li>
                                Scrape, mine, or harvest data from the service
                            </li>
                            <li>
                                Impersonate others or misrepresent your
                                affiliation
                            </li>
                            <li>Transmit viruses, malware, or harmful code</li>
                            <li>
                                Interfere with or disrupt the service or servers
                            </li>
                            <li>
                                Violate any third-party intellectual property
                                rights
                            </li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            7. Intellectual Property Rights
                        </h3>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            7.1 Our Rights
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            Gemmie and all related trademarks, logos, and
                            intellectual property are owned by us or our
                            licensors. You are granted a limited, non-exclusive,
                            non-transferable license to use the service for
                            personal purposes only.
                        </p>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            7.2 Your Content
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            You retain ownership of content you submit. By using
                            Gemmie, you grant us a license to process, store,
                            and display your content solely for providing the
                            service. We may use anonymized data to improve our
                            AI models.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            8. AI-Generated Content
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-2"
                        >
                            AI-generated responses are provided for
                            informational purposes only:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                We do not guarantee accuracy, completeness, or
                                reliability
                            </li>
                            <li>
                                Do not rely on AI responses for medical, legal,
                                or financial advice
                            </li>
                            <li>Verify important information independently</li>
                            <li>
                                We are not liable for decisions made based on AI
                                responses
                            </li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            9. Termination
                        </h3>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            9.1 By You
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            You may terminate your account at any time through
                            the settings page. Upon termination, your data will
                            be deleted within 30 days.
                        </p>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            9.2 By Us
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            We may suspend or terminate your account if you
                            violate these terms, engage in fraudulent activity,
                            or for any reason with 30 days notice (except in
                            cases of serious violation).
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            10. Disclaimers and Limitations of Liability
                        </h3>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            10.1 Service Warranty
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            The service is provided "as is" without warranties
                            of any kind. We do not guarantee uninterrupted or
                            error-free service.
                        </p>

                        <h4
                            class="text-lg font-medium text-gray-900 dark:text-white mt-4 mb-2"
                        >
                            10.2 Limitation of Liability
                        </h4>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            To the maximum extent permitted by Kenyan law, we
                            shall not be liable for indirect, incidental,
                            special, or consequential damages. Our total
                            liability shall not exceed the amount you paid for
                            the service in the last 12 months.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            11. Indemnification
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            You agree to indemnify and hold Gemmie harmless from
                            any claims, damages, or expenses arising from your
                            use of the service, violation of these terms, or
                            infringement of any third-party rights.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            12. Dispute Resolution
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed mb-3"
                        >
                            In accordance with Kenyan law:
                        </p>
                        <ul
                            class="list-disc list-inside text-gray-700 dark:text-gray-300 space-y-1 ml-4"
                        >
                            <li>
                                Disputes shall first be resolved through good
                                faith negotiation
                            </li>
                            <li>
                                If unresolved, disputes may be referred to
                                mediation under the Nairobi Centre for
                                International Arbitration (NCIA)
                            </li>
                            <li>
                                Any legal action must be brought in the courts
                                of Kenya
                            </li>
                            <li>
                                These terms are governed by the laws of Kenya
                            </li>
                        </ul>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            13. Force Majeure
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            We shall not be liable for failure to perform due to
                            circumstances beyond our reasonable control,
                            including natural disasters, acts of government,
                            telecommunications failures, or other force majeure
                            events.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            14. Severability
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            If any provision of these terms is found to be
                            unenforceable under Kenyan law, the remaining
                            provisions shall remain in full force and effect.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            15. Changes to Terms
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            We may modify these terms at any time. Material
                            changes will be notified via email or app
                            notification at least 30 days before taking effect.
                            Continued use after changes constitutes acceptance.
                        </p>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            16. Contact Information
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            For questions about these Terms of Service, contact
                            us at:
                        </p>
                        <div
                            class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg mt-3"
                        >
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Email:</strong> legal@gemmie.app
                            </p>
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Address:</strong> Nairobi, Kenya
                            </p>
                            <p class="text-gray-700 dark:text-gray-300">
                                <strong>Support:</strong> support@gemmie.app
                            </p>
                        </div>
                    </section>

                    <section class="mb-6">
                        <h3
                            class="text-xl font-semibold text-gray-900 dark:text-white mb-3"
                        >
                            17. Entire Agreement
                        </h3>
                        <p
                            class="text-gray-700 dark:text-gray-300 leading-relaxed"
                        >
                            These Terms of Service, together with our Privacy
                            Policy, constitute the entire agreement between you
                            and Gemmie regarding the use of our service.
                        </p>
                    </section>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Smooth scrolling */
html {
    scroll-behavior: smooth;
}

/* Custom prose styling for better readability */
.prose h3 {
    scroll-margin-top: 2rem;
}

.prose a {
    word-break: break-word;
}
</style>

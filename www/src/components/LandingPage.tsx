const LandingPage = () => {
    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-100">
            <div className="bg-white p-8 rounded-lg shadow-lg">
                <h1 className="text-3xl font-semibold text-gray-800 mb-4">
                    Welcome to Betrayal
                </h1>
                <p className="text-gray-600 mb-6">
                    Sign in to access your account
                </p>

                {/* Sign-in Form */}
                <form>
                    <div className="mb-4">
                        <label
                            className="block text-gray-600 text-sm font-semibold mb-2"
                            htmlFor="email"
                        >
                            Email
                        </label>
                        <input
                            className="w-full p-2 border border-gray-300 rounded"
                            type="email"
                            id="email"
                            name="email"
                            placeholder="Your Email"
                            required
                        />
                    </div>

                    <div className="mb-4">
                        <label
                            className="block text-gray-600 text-sm font-semibold mb-2"
                            htmlFor="password"
                        >
                            Password
                        </label>
                        <input
                            className="w-full p-2 border border-gray-300 rounded"
                            type="password"
                            id="password"
                            name="password"
                            placeholder="Your Password"
                            required
                        />
                    </div>

                    <div className="mb-6">
                        <button
                            className="w-full bg-blue-500 hover:bg-blue-600 text-white font-semibold p-3 rounded"
                            type="submit"
                        >
                            Sign In
                        </button>
                    </div>
                </form>

                <p className="text-gray-600 text-sm text-center">
                    Don't have an account?{' '}
                    <a href="/signup" className="text-blue-500">
                        Sign up here
                    </a>
                </p>
            </div>
        </div>
    );
};

export default LandingPage;

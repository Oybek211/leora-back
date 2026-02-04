-- Ensure supported languages exist in meta_languages
INSERT INTO meta_languages (code, name, is_active, is_default) VALUES
  ('en', 'English',  true, true),
  ('ru', 'Русский',  true, false),
  ('uz', 'O''zbek',  true, false),
  ('ar', 'العربية',  true, false),
  ('tr', 'Türkçe',   true, false)
ON CONFLICT (code) DO UPDATE SET is_active = true;

-- Auth & user error translations for all supported languages
INSERT INTO error_translations (code, lang_code, message) VALUES
  -- USER_NOT_FOUND
  ('USER_NOT_FOUND', 'en', 'User not found. Please register first.'),
  ('USER_NOT_FOUND', 'ru', 'Пользователь не найден. Пожалуйста, зарегистрируйтесь.'),
  ('USER_NOT_FOUND', 'uz', 'Foydalanuvchi topilmadi. Iltimos, avval ro''yxatdan o''ting.'),
  ('USER_NOT_FOUND', 'ar', 'المستخدم غير موجود. يرجى التسجيل أولاً.'),
  ('USER_NOT_FOUND', 'tr', 'Kullanıcı bulunamadı. Lütfen önce kayıt olun.'),

  -- USER_ALREADY_EXISTS
  ('USER_ALREADY_EXISTS', 'en', 'This account already exists. Please log in instead.'),
  ('USER_ALREADY_EXISTS', 'ru', 'Этот аккаунт уже существует. Пожалуйста, войдите.'),
  ('USER_ALREADY_EXISTS', 'uz', 'Bu hisob allaqachon mavjud. Iltimos, tizimga kiring.'),
  ('USER_ALREADY_EXISTS', 'ar', 'هذا الحساب موجود بالفعل. يرجى تسجيل الدخول بدلاً من ذلك.'),
  ('USER_ALREADY_EXISTS', 'tr', 'Bu hesap zaten mevcut. Lütfen giriş yapın.'),

  -- INVALID_CREDENTIALS
  ('INVALID_CREDENTIALS', 'en', 'Invalid email or password.'),
  ('INVALID_CREDENTIALS', 'ru', 'Неверный email или пароль.'),
  ('INVALID_CREDENTIALS', 'uz', 'Email yoki parol noto''g''ri.'),
  ('INVALID_CREDENTIALS', 'ar', 'البريد الإلكتروني أو كلمة المرور غير صحيحة.'),
  ('INVALID_CREDENTIALS', 'tr', 'Geçersiz e-posta veya şifre.'),

  -- INVALID_TOKEN
  ('INVALID_TOKEN', 'en', 'Your session has expired. Please log in again.'),
  ('INVALID_TOKEN', 'ru', 'Ваша сессия истекла. Пожалуйста, войдите снова.'),
  ('INVALID_TOKEN', 'uz', 'Sessiya muddati tugadi. Iltimos, qaytadan kiring.'),
  ('INVALID_TOKEN', 'ar', 'انتهت صلاحية جلستك. يرجى تسجيل الدخول مرة أخرى.'),
  ('INVALID_TOKEN', 'tr', 'Oturumunuz sona erdi. Lütfen tekrar giriş yapın.'),

  -- INVALID_GOOGLE_TOKEN
  ('INVALID_GOOGLE_TOKEN', 'en', 'Google authentication failed. Please try again.'),
  ('INVALID_GOOGLE_TOKEN', 'ru', 'Ошибка аутентификации Google. Попробуйте снова.'),
  ('INVALID_GOOGLE_TOKEN', 'uz', 'Google autentifikatsiyasi muvaffaqiyatsiz. Qaytadan urinib ko''ring.'),
  ('INVALID_GOOGLE_TOKEN', 'ar', 'فشلت مصادقة Google. يرجى المحاولة مرة أخرى.'),
  ('INVALID_GOOGLE_TOKEN', 'tr', 'Google kimlik doğrulaması başarısız. Lütfen tekrar deneyin.'),

  -- INVALID_APPLE_TOKEN
  ('INVALID_APPLE_TOKEN', 'en', 'Apple authentication failed. Please try again.'),
  ('INVALID_APPLE_TOKEN', 'ru', 'Ошибка аутентификации Apple. Попробуйте снова.'),
  ('INVALID_APPLE_TOKEN', 'uz', 'Apple autentifikatsiyasi muvaffaqiyatsiz. Qaytadan urinib ko''ring.'),
  ('INVALID_APPLE_TOKEN', 'ar', 'فشلت مصادقة Apple. يرجى المحاولة مرة أخرى.'),
  ('INVALID_APPLE_TOKEN', 'tr', 'Apple kimlik doğrulaması başarısız. Lütfen tekrar deneyin.'),

  -- INVALID_USER_DATA
  ('INVALID_USER_DATA', 'en', 'Invalid user data. Please check your input.'),
  ('INVALID_USER_DATA', 'ru', 'Неверные данные. Пожалуйста, проверьте введённые данные.'),
  ('INVALID_USER_DATA', 'uz', 'Noto''g''ri ma''lumotlar. Iltimos, kiritilgan ma''lumotlarni tekshiring.'),
  ('INVALID_USER_DATA', 'ar', 'بيانات المستخدم غير صالحة. يرجى التحقق من المدخلات.'),
  ('INVALID_USER_DATA', 'tr', 'Geçersiz kullanıcı verileri. Lütfen girişinizi kontrol edin.'),

  -- PERMISSION_DENIED
  ('PERMISSION_DENIED', 'en', 'You do not have permission to perform this action.'),
  ('PERMISSION_DENIED', 'ru', 'У вас нет разрешения для выполнения этого действия.'),
  ('PERMISSION_DENIED', 'uz', 'Sizda bu amalni bajarish uchun ruxsat yo''q.'),
  ('PERMISSION_DENIED', 'ar', 'ليس لديك إذن للقيام بهذا الإجراء.'),
  ('PERMISSION_DENIED', 'tr', 'Bu işlemi gerçekleştirmek için izniniz yok.')

ON CONFLICT (code, lang_code) DO UPDATE SET message = EXCLUDED.message;

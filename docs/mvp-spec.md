# MVP Spec – Create Booking (POST /bookings)

Bu endpoint istifadəçiyə biznes üçün müəyyən tarix və saat aralığında booking yaratmağa imkan verir. Qaydalar, cavab strukturu və reminder job yaratma davranışı aşağıda müəyyən olunub.

# Use-case

İstifadəçi müəyyən biznes üçün tarix və saat aralığı seçərək booking yaratmaq istəyir.

Sistem yalnız gələcək tarixlər üçün və boş olan slotlarda booking yaradılmasına icazə verir.

Uğurlu booking yaradıldıqda, istifadəçiyə xatırlatma göndərmək üçün daxili reminder job planlaşdırılır.

# Endpoint

Method: POST

Path: /bookings

Auth: (burada sisteminə görə yazırsan: məsələn, “Authenticated user (Bearer token)” və ya “TBD”)

Request format: Content-Type: application/json

Response format: application/json

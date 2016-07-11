SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- Database: `bitanchor`
--

-- --------------------------------------------------------

--
-- Table structure for table `records`
--

CREATE TABLE `records` (
  `id` int(11) NOT NULL,
  `account` text NOT NULL,
  `wallet` text NOT NULL,
  `outgoing_wallet` text NOT NULL,
  `return_wallet` text NOT NULL,
  `amount` float NOT NULL,
  `fee` float NOT NULL,
  `callback_url` text NOT NULL,
  `notify_method` varchar(16) NOT NULL,
  `notify_value` text NOT NULL,
  `password` text NOT NULL,
  `active` tinyint(1) NOT NULL,
  `paid` tinyint(1) NOT NULL DEFAULT '0',
  `sent` tinyint(1) NOT NULL DEFAULT '0',
  `locked` tinyint(1) NOT NULL,
  `transaction_id` text NOT NULL,
  `out_transaction_id` text NOT NULL,
  `qrcode_file` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `records`
--
ALTER TABLE `records`
  ADD PRIMARY KEY (`id`);
